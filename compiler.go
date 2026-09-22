package main

// compiler.go - Ahead-of-time compiler from a loaded VASM program to LLVM IR.
//
// The VM in this repository is a dynamically typed stack machine: every slot of
// the stack is a `Value_Holder` holding an int64, a float64 and a string, and
// the *type* of a slot is recovered at runtime by comparing the int64/float64
// fields against sentinel values (see get_operand_type_by_name). There is no
// static type information in a .vasm program, so a faithful compiler cannot
// erase the Value_Holder representation.
//
// The strategy is therefore:
//
//   * the VM stack becomes a global array of `%vh = { i64, double, i8* }`,
//     initialised at startup to the same sentinels the Go VM uses
//     ({MinInt64, SmallestNonzeroFloat64, ""});
//   * every VM instruction becomes its own LLVM basic block, so jumps that are
//     statically known (JMP / JMP_IF / CALL) become real `br` edges instead of
//     an interpreter dispatch;
//   * RET pops a return address that is only known at runtime, so it goes
//     through an `indirectbr` over a table of `blockaddress` constants;
//   * every runtime check, every error message and every printing quirk of the
//     Go implementation is reproduced, including the ones that look like bugs
//     (they are behaviour this VM's programs already depend on). Each such case
//     is called out with a comment below.
//
// The emitted module only depends on the C standard library (printf, puts,
// putchar, snprintf, strtod, strchr, atoi, exit), so it can be built with
// `clang out.ll -o out` on any platform LLVM supports.

import (
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// ---------------------------------------------------------------------------
// Sentinels shared with the interpreter
// ---------------------------------------------------------------------------

// LLVM spells double constants as the 16 hex digits of their IEEE-754 bits.
func llvmDouble(f float64) string {
	return fmt.Sprintf("0x%016X", math.Float64bits(f))
}

// ---------------------------------------------------------------------------
// Emitter
// ---------------------------------------------------------------------------

type llvm_compiler struct {
	body strings.Builder

	reg_counter int

	// Interned string literals. `$S[[text]]` anywhere in the emitted text is
	// replaced by an inline getelementptr to a private constant holding `text`.
	str_names map[string]string
	str_order []string

	program_size int64
	limit        int
	limited      bool
	needs_iptbl  bool
}

func (c *llvm_compiler) w(format string, args ...interface{}) {
	fmt.Fprintf(&c.body, format, args...)
	c.body.WriteByte('\n')
}

// reg returns a fresh SSA register name. vm.main is a single function, so the
// counter is global to the whole emission.
func (c *llvm_compiler) reg() string {
	c.reg_counter += 1
	return fmt.Sprintf("%%v%d", c.reg_counter)
}

var llvm_str_token = regexp.MustCompile(`\$S\[\[((?s).*?)\]\]`)

// llvm_unescape expands the handful of escapes accepted inside a $S[[..]] token.
func llvm_unescape(s string) string {
	var out strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case 'n':
				out.WriteByte('\n')
				i += 1
				continue
			case 't':
				out.WriteByte('\t')
				i += 1
				continue
			case '\\':
				out.WriteByte('\\')
				i += 1
				continue
			}
		}
		out.WriteByte(s[i])
	}
	return out.String()
}

// llvm_quote renders a Go string as the body of an LLVM `c"..."` literal.
func llvm_quote(s string) string {
	var out strings.Builder
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '"' || ch == '\\' || ch < 0x20 || ch > 0x7E {
			fmt.Fprintf(&out, "\\%02X", ch)
		} else {
			out.WriteByte(ch)
		}
	}
	return out.String()
}

// intern registers a string literal and returns the global's symbol name.
func (c *llvm_compiler) intern(text string) string {
	if name, ok := c.str_names[text]; ok {
		return name
	}
	name := fmt.Sprintf("@.vm.str.%d", len(c.str_order))
	c.str_names[text] = name
	c.str_order = append(c.str_order, text)
	return name
}

// cstr interns `text` verbatim and returns the inline constant expression that
// points at its first byte. Used for operands that come from the source program,
// which must not be run through the $S[[..]]/$TOKEN substitution passes.
func (c *llvm_compiler) cstr(text string) string {
	name := c.intern(text)
	n := len(text) + 1
	return fmt.Sprintf("getelementptr inbounds ([%d x i8], [%d x i8]* %s, i64 0, i64 0)", n, n, name)
}

// resolve_strings performs the $S[[..]] substitution pass over emitted text.
func (c *llvm_compiler) resolve_strings(text string) string {
	return llvm_str_token.ReplaceAllStringFunc(text, func(match string) string {
		groups := llvm_str_token.FindStringSubmatch(match)
		literal := llvm_unescape(groups[1])
		name := c.intern(literal)
		n := len(literal) + 1
		return fmt.Sprintf("getelementptr inbounds ([%d x i8], [%d x i8]* %s, i64 0, i64 0)", n, n, name)
	})
}

func (c *llvm_compiler) string_globals() string {
	var out strings.Builder
	for i, text := range c.str_order {
		fmt.Fprintf(&out, "@.vm.str.%d = private unnamed_addr constant [%d x i8] c\"%s\\00\"\n",
			i, len(text)+1, llvm_quote(text))
	}
	return out.String()
}

// ---------------------------------------------------------------------------
// Addressing helpers
// ---------------------------------------------------------------------------

// slot emits a getelementptr to field `field` of STACK[index] and returns the
// register holding that pointer. field: 0 = int64holder, 1 = float64holder,
// 2 = pointer.
func (c *llvm_compiler) slot(index string, field int) string {
	p := c.reg()
	c.w("  %s = getelementptr inbounds [%d x %%vh], [%d x %%vh]* @.vm.stack, i64 0, i64 %s, i32 %d",
		p, STACK_CAPACITY, STACK_CAPACITY, index, field)
	return p
}

func (c *llvm_compiler) load_i(index string) string {
	p := c.slot(index, 0)
	v := c.reg()
	c.w("  %s = load i64, i64* %s, align 8", v, p)
	return v
}

func (c *llvm_compiler) load_f(index string) string {
	p := c.slot(index, 1)
	v := c.reg()
	c.w("  %s = load double, double* %s, align 8", v, p)
	return v
}

func (c *llvm_compiler) load_ss() string {
	v := c.reg()
	c.w("  %s = load i64, i64* @.vm.ss, align 8", v)
	return v
}

func (c *llvm_compiler) store_ss(value string) {
	c.w("  store i64 %s, i64* @.vm.ss, align 8", value)
}

func (c *llvm_compiler) addk(value string, k int64) string {
	r := c.reg()
	c.w("  %s = add nsw i64 %s, %d", r, value, k)
	return r
}

// ---------------------------------------------------------------------------
// Runtime support library (hand written, target independent LLVM IR)
// ---------------------------------------------------------------------------

const llvm_runtime_template = `
; --- libc ------------------------------------------------------------------
declare i32 @printf(i8*, ...)
declare i32 @puts(i8*)
declare i32 @putchar(i32)
declare i32 @snprintf(i8*, i64, i8*, ...)
declare double @strtod(i8*, i8**)
declare i8* @strchr(i8*, i32)
declare i32 @atoi(i8*)
declare void @exit(i32)

; --- machine state ----------------------------------------------------------
; A Value_Holder: { int64holder, float64holder, pointer }.
%vh = type { i64, double, i8* }

@.vm.stack = internal global [$CAP x %vh] zeroinitializer, align 8
@.vm.ss    = internal global i64 0, align 8
@.vm.steps = internal global i64 0, align 8
@.vm.ip    = internal global i64 0, align 8

; init_all(): every stack slot starts at {MinInt64, SmallestNonzeroFloat64, ""}.
; Those are the sentinels get_operand_type_by_name() tests against, so the
; initial value of a slot is "no type at all".
define internal void @vm.init() {
entry:
  br label %loop
loop:
  %i = phi i64 [ 0, %entry ], [ %next, %loop ]
  %pi = getelementptr inbounds [$CAP x %vh], [$CAP x %vh]* @.vm.stack, i64 0, i64 %i, i32 0
  store i64 $MININT, i64* %pi, align 8
  %pf = getelementptr inbounds [$CAP x %vh], [$CAP x %vh]* @.vm.stack, i64 0, i64 %i, i32 1
  store double $MINFLOAT, double* %pf, align 8
  %pp = getelementptr inbounds [$CAP x %vh], [$CAP x %vh]* @.vm.stack, i64 0, i64 %i, i32 2
  store i8* null, i8** %pp, align 8
  %next = add nuw nsw i64 %i, 1
  %more = icmp slt i64 %next, $CAP
  br i1 %more, label %loop, label %done
done:
  ret void
}

; exit_with_one(): the Go VM prints the message on *stdout* and exits 1.
define internal void @vm.fatal(i8* %msg) {
entry:
  %0 = call i32 @puts(i8* %msg)
  call void @exit(i32 1)
  unreachable
}

; A Go panic. The real runtime writes to stderr; we keep everything on stdout so
; the emitted module needs no platform specific FILE* globals.
define internal void @vm.panic(i8* %msg) {
entry:
  %0 = call i32 (i8*, ...) @printf(i8* $S[[panic: %s\n]], i8* %msg)
  call void @exit(i32 2)
  unreachable
}

; Go's "index out of range" panic, reachable the same way it is in the VM.
define internal void @vm.oob(i64 %idx) {
entry:
  %neg = icmp slt i64 %idx, 0
  br i1 %neg, label %short, label %long
short:
  %0 = call i32 (i8*, ...) @printf(i8* $S[[panic: runtime error: index out of range [%lld]\n]], i64 %idx)
  br label %fin
long:
  %1 = call i32 (i8*, ...) @printf(i8* $S[[panic: runtime error: index out of range [%lld] with length $PROGCAP\n]], i64 %idx)
  br label %fin
fin:
  call void @exit(i32 2)
  unreachable
}

; execute_inst() prologue: "Illegal Instruction Access". The trace line prints
; the zero value of Inst, which is what PROGRAM[ip] holds past program_size.
define internal void @vm.illegal() {
entry:
  %0 = call i32 @puts(i8* $S[[Instruction :  : {$MININT $MINFLOAT_G }]])
  call void @vm.fatal(i8* $S[[Illegal Instruction Access]])
  unreachable
}

; execute_inst() prologue: the two stack sanity checks run before every
; instruction.
define internal void @vm.guard() {
entry:
  %ss = load i64, i64* @.vm.ss, align 8
  %under = icmp slt i64 %ss, 0
  br i1 %under, label %uf, label %chk_over
uf:
  call void @vm.fatal(i8* $S[[Stack Underflow]])
  unreachable
chk_over:
  %over = icmp sgt i64 %ss, $CAP
  br i1 %over, label %of, label %ok
of:
  call void @vm.fatal(i8* $S[[Stack Overflow]])
  unreachable
ok:
  ret void
}

; get_operand_type_by_name(): 0 = int64, 1 = float64, 2 = pointer.
; The probe order (float, then int, then pointer) is significant and preserved.
define internal i32 @vm.typeof(i64 %idx) {
entry:
  %pf = getelementptr inbounds [$CAP x %vh], [$CAP x %vh]* @.vm.stack, i64 0, i64 %idx, i32 1
  %f = load double, double* %pf, align 8
  %notminf = fcmp une double %f, $MINFLOAT
  br i1 %notminf, label %isfloat, label %chkint
isfloat:
  ret i32 1
chkint:
  %pi = getelementptr inbounds [$CAP x %vh], [$CAP x %vh]* @.vm.stack, i64 0, i64 %idx, i32 0
  %i = load i64, i64* %pi, align 8
  %notmini = icmp ne i64 %i, $MININT
  br i1 %notmini, label %isint, label %chkptr
isint:
  ret i32 0
chkptr:
  %pp = getelementptr inbounds [$CAP x %vh], [$CAP x %vh]* @.vm.stack, i64 0, i64 %idx, i32 2
  %p = load i8*, i8** %pp, align 8
  %isnull = icmp eq i8* %p, null
  br i1 %isnull, label %unreach, label %chkempty
chkempty:
  %c0 = load i8, i8* %p, align 1
  %isempty = icmp eq i8 %c0, 0
  br i1 %isempty, label %unreach, label %isptr
isptr:
  ret i32 2
unreach:
  call void @vm.panic(i8* $S[[Unreachable]])
  unreachable
}

; The stack-depth precondition shared by most instructions: error out when
; stack_size < n. Returns the current stack_size.
define internal i64 @vm.need(i64 %n, i8* %msg) {
entry:
  %ss = load i64, i64* @.vm.ss, align 8
  %bad = icmp slt i64 %ss, %n
  br i1 %bad, label %fail, label %ok
fail:
  call void @vm.fatal(i8* %msg)
  unreachable
ok:
  ret i64 %ss
}

; The type precondition of the binary operators: both of the top two slots must
; have type %want. The check short circuits exactly like the Go $BT&&$BT does, and
; dumps the stack before dying, like the Go code.
define internal void @vm.require2(i32 %want, i8* %msg) {
entry:
  %ss = load i64, i64* @.vm.ss, align 8
  %a = add nsw i64 %ss, -1
  %b = add nsw i64 %ss, -2
  %ta = call i32 @vm.typeof(i64 %a)
  %oka = icmp eq i32 %ta, %want
  br i1 %oka, label %chkb, label %fail
chkb:
  %tb = call i32 @vm.typeof(i64 %b)
  %okb = icmp eq i32 %tb, %want
  br i1 %okb, label %ok, label %fail
fail:
  call void @vm.print_stack(i1 true)
  call void @vm.fatal(i8* %msg)
  unreachable
ok:
  ret void
}

; ---------------------------------------------------------------------------
; Formatting. Go's fmt uses shortest-round-trip decimal for %v on a float64,
; which C's printf cannot do directly, so vm.gfmt reproduces
; strconv.FormatFloat(v, 'g', -1, 64):
;   * find the smallest digit count D in 1..17 that round-trips through strtod;
;   * let exp be the decimal exponent of that representation;
;   * use %e when exp < -4 || exp >= 6, %f otherwise (Go clamps the 'g'
;     crossover precision to 6 when the precision is "shortest").
; %buf must have room for at least 48 bytes.
; ---------------------------------------------------------------------------
define internal void @vm.gfmt(double %v, i8* %buf) {
entry:
  %isnan = fcmp uno double %v, %v
  br i1 %isnan, label %nan, label %chkpinf
nan:
  %0 = call i32 (i8*, i64, i8*, ...) @snprintf(i8* %buf, i64 48, i8* $S[[%s]], i8* $S[[NaN]])
  ret void
chkpinf:
  %ispinf = fcmp oeq double %v, 0x7FF0000000000000
  br i1 %ispinf, label %pinf, label %chkninf
pinf:
  %1 = call i32 (i8*, i64, i8*, ...) @snprintf(i8* %buf, i64 48, i8* $S[[%s]], i8* $S[[+Inf]])
  ret void
chkninf:
  %isninf = fcmp oeq double %v, 0xFFF0000000000000
  br i1 %isninf, label %ninf, label %probe
ninf:
  %2 = call i32 (i8*, i64, i8*, ...) @snprintf(i8* %buf, i64 48, i8* $S[[%s]], i8* $S[[-Inf]])
  ret void
probe:
  br label %probe.head
probe.head:
  %d = phi i64 [ 1, %probe ], [ %dnext, %probe.next ]
  %prec = add nsw i64 %d, -1
  %prec32 = trunc i64 %prec to i32
  %3 = call i32 (i8*, i64, i8*, ...) @snprintf(i8* %buf, i64 48, i8* $S[[%.*e]], i32 %prec32, double %v)
  %back = call double @strtod(i8* %buf, i8** null)
  %same = fcmp oeq double %back, %v
  br i1 %same, label %found, label %probe.next
probe.next:
  %dnext = add nuw nsw i64 %d, 1
  %again = icmp sle i64 %dnext, 17
  br i1 %again, label %probe.head, label %exhausted
exhausted:
  br label %found
found:
  %digits = phi i64 [ %d, %probe.head ], [ 17, %exhausted ]
  %epos = call i8* @strchr(i8* %buf, i32 101)
  %noexp = icmp eq i8* %epos, null
  br i1 %noexp, label %done, label %haveexp
haveexp:
  %edigits = getelementptr inbounds i8, i8* %epos, i64 1
  %exp32 = call i32 @atoi(i8* %edigits)
  %exp = sext i32 %exp32 to i64
  %small = icmp slt i64 %exp, -4
  %large = icmp sge i64 %exp, 6
  %sci = or i1 %small, %large
  br i1 %sci, label %done, label %fixed
fixed:
  %t0 = add nsw i64 %digits, -1
  %t1 = sub nsw i64 %t0, %exp
  %isneg = icmp slt i64 %t1, 0
  %fprec = select i1 %isneg, i64 0, i64 %t1
  %fprec32 = trunc i64 %fprec to i32
  %4 = call i32 (i8*, i64, i8*, ...) @snprintf(i8* %buf, i64 48, i8* $S[[%.*f]], i32 %fprec32, double %v)
  br label %done
done:
  ret void
}

; fmt.Printf("%f\n", v). Go spells the specials "+Inf" / "-Inf" / "NaN".
define internal void @vm.printf_float(double %v) {
entry:
  %isnan = fcmp uno double %v, %v
  br i1 %isnan, label %nan, label %chkpinf
nan:
  %0 = call i32 @puts(i8* $S[[NaN]])
  ret void
chkpinf:
  %ispinf = fcmp oeq double %v, 0x7FF0000000000000
  br i1 %ispinf, label %pinf, label %chkninf
pinf:
  %1 = call i32 @puts(i8* $S[[+Inf]])
  ret void
chkninf:
  %isninf = fcmp oeq double %v, 0xFFF0000000000000
  br i1 %isninf, label %ninf, label %plain
ninf:
  %2 = call i32 @puts(i8* $S[[-Inf]])
  ret void
plain:
  %3 = call i32 (i8*, ...) @printf(i8* $S[[%f\n]], double %v)
  ret void
}

; fmt.Println(Value_Holder{...}) -> "{<int> <float> <string>}".
define internal void @vm.print_vh(i64 %idx) {
entry:
  %buf = alloca [48 x i8], align 1
  %bp = getelementptr inbounds [48 x i8], [48 x i8]* %buf, i64 0, i64 0
  %pi = getelementptr inbounds [$CAP x %vh], [$CAP x %vh]* @.vm.stack, i64 0, i64 %idx, i32 0
  %i = load i64, i64* %pi, align 8
  %pf = getelementptr inbounds [$CAP x %vh], [$CAP x %vh]* @.vm.stack, i64 0, i64 %idx, i32 1
  %f = load double, double* %pf, align 8
  %pp = getelementptr inbounds [$CAP x %vh], [$CAP x %vh]* @.vm.stack, i64 0, i64 %idx, i32 2
  %p = load i8*, i8** %pp, align 8
  call void @vm.gfmt(double %f, i8* %bp)
  %isnull = icmp eq i8* %p, null
  %pstr = select i1 %isnull, i8* $S[[]], i8* %p
  %0 = call i32 (i8*, ...) @printf(i8* $S[[{%lld %s %s}\n]], i64 %i, i8* %bp, i8* %pstr)
  ret void
}

; print_stack()
define internal void @vm.print_stack(i1 %reverse) {
entry:
  %ss = load i64, i64* @.vm.ss, align 8
  %under = icmp slt i64 %ss, 0
  br i1 %under, label %uf, label %ok
uf:
  call void @vm.fatal(i8* $S[[ERROR: Stack Underflow]])
  unreachable
ok:
  %0 = call i32 @puts(i8* $S[[---- STACK BEG ----]])
  %top = add nsw i64 %ss, -1
  br i1 %reverse, label %rev.head, label %fwd.head
rev.head:
  %ri = phi i64 [ %top, %ok ], [ %rnext, %rev.body ]
  %rgo = icmp sge i64 %ri, 0
  br i1 %rgo, label %rev.body, label %tail
rev.body:
  call void @vm.print_vh(i64 %ri)
  %rnext = add nsw i64 %ri, -1
  br label %rev.head
fwd.head:
  %fi = phi i64 [ 0, %ok ], [ %fnext, %fwd.body ]
  %fgo = icmp slt i64 %fi, %ss
  br i1 %fgo, label %fwd.body, label %tail
fwd.body:
  call void @vm.print_vh(i64 %fi)
  %fnext = add nsw i64 %fi, 1
  br label %fwd.head
tail:
  %1 = call i32 @puts(i8* $S[[---- STACK END ----]])
  %2 = call i32 @puts(i8* $S[[]])
  ret void
}

; print(): dispatch on the runtime type of the top of the stack.
define internal void @vm.emit_value(i64 %idx) {
entry:
  %t = call i32 @vm.typeof(i64 %idx)
  switch i32 %t, label %as_ptr [
    i32 0, label %as_int
    i32 1, label %as_float
  ]
as_int:
  %pi = getelementptr inbounds [$CAP x %vh], [$CAP x %vh]* @.vm.stack, i64 0, i64 %idx, i32 0
  %i = load i64, i64* %pi, align 8
  %0 = call i32 (i8*, ...) @printf(i8* $S[[%lld\n]], i64 %i)
  ret void
as_float:
  %pf = getelementptr inbounds [$CAP x %vh], [$CAP x %vh]* @.vm.stack, i64 0, i64 %idx, i32 1
  %f = load double, double* %pf, align 8
  call void @vm.printf_float(double %f)
  ret void
as_ptr:
  %pp = getelementptr inbounds [$CAP x %vh], [$CAP x %vh]* @.vm.stack, i64 0, i64 %idx, i32 2
  %p = load i8*, i8** %pp, align 8
  %isnull = icmp eq i8* %p, null
  %pstr = select i1 %isnull, i8* $S[[]], i8* %p
  %1 = call i32 (i8*, ...) @printf(i8* $S[[%s\n]], i8* %pstr)
  ret void
}

; print_asc(): Go's string(rune(n)) - UTF-8 encode, U+FFFD for anything that is
; not a valid scalar value. Bytes go out one at a time so an embedded NUL is
; written faithfully.
define internal void @vm.print_asc(i64 %idx) {
entry:
  %t = call i32 @vm.typeof(i64 %idx)
  %isint = icmp eq i32 %t, 0
  br i1 %isint, label %encode, label %wrongtype
wrongtype:
  %isfloat = icmp eq i32 %t, 1
  %tname = select i1 %isfloat, i8* $S[[float64]], i8* $S[[pointer]]
  %0 = call i32 (i8*, ...) @printf(i8* $S[[\nERROR: Runtime error: Instruction $BTPRINT_ASC$BT failed: Expected type $BTinteger$BT on top of stack but found $BT%s$BT\n]], i8* %tname)
  ret void
encode:
  %pi = getelementptr inbounds [$CAP x %vh], [$CAP x %vh]* @.vm.stack, i64 0, i64 %idx, i32 0
  %raw = load i64, i64* %pi, align 8
  %neg = icmp slt i64 %raw, 0
  %big = icmp sgt i64 %raw, 1114111
  %sur_lo = icmp sge i64 %raw, 55296
  %sur_hi = icmp sle i64 %raw, 57343
  %sur = and i1 %sur_lo, %sur_hi
  %bad0 = or i1 %neg, %big
  %bad = or i1 %bad0, %sur
  %cp = select i1 %bad, i64 65533, i64 %raw
  %one = icmp slt i64 %cp, 128
  br i1 %one, label %enc1, label %chk2
enc1:
  %b0 = trunc i64 %cp to i32
  %1 = call i32 @putchar(i32 %b0)
  ret void
chk2:
  %two = icmp slt i64 %cp, 2048
  br i1 %two, label %enc2, label %chk3
enc2:
  %s6 = lshr i64 %cp, 6
  %h2 = or i64 %s6, 192
  %h2i = trunc i64 %h2 to i32
  %2 = call i32 @putchar(i32 %h2i)
  %l2 = and i64 %cp, 63
  %l2o = or i64 %l2, 128
  %l2i = trunc i64 %l2o to i32
  %3 = call i32 @putchar(i32 %l2i)
  ret void
chk3:
  %three = icmp slt i64 %cp, 65536
  br i1 %three, label %enc3, label %enc4
enc3:
  %t12 = lshr i64 %cp, 12
  %t12o = or i64 %t12, 224
  %t12i = trunc i64 %t12o to i32
  %4 = call i32 @putchar(i32 %t12i)
  %t6 = lshr i64 %cp, 6
  %t6m = and i64 %t6, 63
  %t6o = or i64 %t6m, 128
  %t6i = trunc i64 %t6o to i32
  %5 = call i32 @putchar(i32 %t6i)
  %t0 = and i64 %cp, 63
  %t0o = or i64 %t0, 128
  %t0i = trunc i64 %t0o to i32
  %6 = call i32 @putchar(i32 %t0i)
  ret void
enc4:
  %q18 = lshr i64 %cp, 18
  %q18o = or i64 %q18, 240
  %q18i = trunc i64 %q18o to i32
  %7 = call i32 @putchar(i32 %q18i)
  %q12 = lshr i64 %cp, 12
  %q12m = and i64 %q12, 63
  %q12o = or i64 %q12m, 128
  %q12i = trunc i64 %q12o to i32
  %8 = call i32 @putchar(i32 %q12i)
  %q6 = lshr i64 %cp, 6
  %q6m = and i64 %q6, 63
  %q6o = or i64 %q6m, 128
  %q6i = trunc i64 %q6o to i32
  %9 = call i32 @putchar(i32 %q6i)
  %q0 = and i64 %cp, 63
  %q0o = or i64 %q0, 128
  %q0i = trunc i64 %q0o to i32
  %10 = call i32 @putchar(i32 %q0i)
  ret void
}
`

// ---------------------------------------------------------------------------
// Instruction emission
// ---------------------------------------------------------------------------

// block returns the label of the basic block holding instruction ip.
func inst_label(ip int64) string { return fmt.Sprintf("ip.%d", ip) }

// branch_to emits the terminator for a statically known transfer of control to
// `target`. Out of range targets reproduce what the interpreter does when
// inst_ptr lands outside the program.
func (c *llvm_compiler) branch_to(target int64) {
	switch {
	case target < 0:
		c.w("  call void @vm.oob(i64 %d)", target)
		c.w("  unreachable")
	case target >= c.program_size:
		// PROGRAM[ip] is the zero Inst, so execute_inst() reports an illegal
		// instruction access.
		c.w("  br label %%%s", inst_label(c.program_size))
	default:
		c.w("  br label %%%s", inst_label(target))
	}
}

// emit_binop_int emits ADDI/SUBI/MULI/DIVI/EQI.
func (c *llvm_compiler) emit_binop_int(op string, need_msg string) {
	c.w("  %s = call i64 @vm.need(i64 2, i8* $S[[%s]])", c.reg(), need_msg)

	if op == "div" {
		// divi() checks for a zero divisor *before* it validates the types.
		ss := c.load_ss()
		a := c.addk(ss, -1)
		div := c.load_i(a)
		iszero := c.reg()
		c.w("  %s = icmp eq i64 %s, 0", iszero, div)
		lz, lnz := c.reg(), c.reg()
		zl, nzl := strings.TrimPrefix(lz, "%")+".zerodiv", strings.TrimPrefix(lnz, "%")+".ok"
		c.w("  br i1 %s, label %%%s, label %%%s", iszero, zl, nzl)
		c.w("%s:", zl)
		c.w("  call void @vm.print_stack(i1 true)")
		c.w("  call void @vm.fatal(i8* $S[[Zero Division Error]])")
		c.w("  unreachable")
		c.w("%s:", nzl)
	}

	c.w("  call void @vm.require2(i32 0, i8* $S[[%s]])", int_type_msg(op))

	ss := c.load_ss()
	a := c.addk(ss, -1)
	b := c.addk(ss, -2)
	va := c.load_i(a)
	pb := c.slot(b, 0)
	vb := c.reg()
	c.w("  %s = load i64, i64* %s, align 8", vb, pb)

	r := c.reg()
	switch op {
	case "add":
		// addi: STACK[ss-1] + STACK[ss-2]
		c.w("  %s = add i64 %s, %s", r, va, vb)
	case "sub":
		c.w("  %s = sub i64 %s, %s", r, vb, va)
	case "mul":
		c.w("  %s = mul i64 %s, %s", r, vb, va)
	case "div":
		c.w("  %s = sdiv i64 %s, %s", r, vb, va)
	case "eq":
		cmp := c.reg()
		c.w("  %s = icmp eq i64 %s, %s", cmp, vb, va)
		c.w("  %s = select i1 %s, i64 1, i64 0", r, cmp)
	}
	c.w("  store i64 %s, i64* %s, align 8", r, pb)
	c.store_ss(a)
}

func int_type_msg(op string) string {
	if op == "eq" {
		return "Invalid Type: Explicitly Push Operands as Int for Integer Equality"
	}
	return "Invalid Type: Explicitly Push Operands as Int for Integer Operands"
}

// emit_binop_float emits ADDF/SUBF/MULF/DIVF/EQF.
func (c *llvm_compiler) emit_binop_float(op string, need_msg string) {
	c.w("  %s = call i64 @vm.need(i64 2, i8* $S[[%s]])", c.reg(), need_msg)

	if op == "div" {
		ss := c.load_ss()
		a := c.addk(ss, -1)
		div := c.load_f(a)
		iszero := c.reg()
		c.w("  %s = fcmp oeq double %s, 0.000000e+00", iszero, div)
		lz, lnz := c.reg(), c.reg()
		zl, nzl := strings.TrimPrefix(lz, "%")+".zerodiv", strings.TrimPrefix(lnz, "%")+".ok"
		c.w("  br i1 %s, label %%%s, label %%%s", iszero, zl, nzl)
		c.w("%s:", zl)
		c.w("  call void @vm.print_stack(i1 true)")
		c.w("  call void @vm.fatal(i8* $S[[Zero Division Error]])")
		c.w("  unreachable")
		c.w("%s:", nzl)
	}

	msg := "Invalid Type: 'Implicit Conversion' to Float Not Yet Supported. Explicitly Push Operands as Float"
	if op == "eq" {
		msg = "Invalid Type: Explicitly Push Operands as Float for Float Equality"
	}
	c.w("  call void @vm.require2(i32 1, i8* $S[[%s]])", msg)

	ss := c.load_ss()
	a := c.addk(ss, -1)
	b := c.addk(ss, -2)
	va := c.load_f(a)
	pb := c.slot(b, 1)
	vb := c.reg()
	c.w("  %s = load double, double* %s, align 8", vb, pb)

	if op == "eq" {
		// eqf() calls reset_operand_except(), which is a no-op in the Go code
		// (it rebinds a local). The float field therefore survives and the slot
		// keeps reporting type float64 even though the result is an integer.
		cmp := c.reg()
		c.w("  %s = fcmp oeq double %s, %s", cmp, vb, va)
		r := c.reg()
		c.w("  %s = select i1 %s, i64 1, i64 0", r, cmp)
		pbi := c.slot(b, 0)
		c.w("  store i64 %s, i64* %s, align 8", r, pbi)
		c.store_ss(a)
		return
	}

	r := c.reg()
	switch op {
	case "add":
		c.w("  %s = fadd double %s, %s", r, va, vb)
	case "sub":
		c.w("  %s = fsub double %s, %s", r, vb, va)
	case "mul":
		c.w("  %s = fmul double %s, %s", r, vb, va)
	case "div":
		c.w("  %s = fdiv double %s, %s", r, vb, va)
	}
	c.w("  store double %s, double* %s, align 8", r, pb)
	c.store_ss(a)
}

// emit_instruction lowers one VM instruction into its basic block.
func (c *llvm_compiler) emit_instruction(ip int64, inst Inst) {
	next := ip + 1

	c.w("")
	c.w("; [%d] %s", ip, describe_inst(inst))
	c.w("%s:", inst_label(ip))

	body_label := inst_label(ip)
	if c.limited {
		// execute_program(): `for vm_halt != 1 && counter < limit`.
		body_label = fmt.Sprintf("ip.%d.body", ip)
		steps := c.reg()
		c.w("  %s = load i64, i64* @.vm.steps, align 8", steps)
		ok := c.reg()
		c.w("  %s = icmp slt i64 %s, %d", ok, steps, c.limit)
		c.w("  br i1 %s, label %%%s, label %%vm.exit", ok, body_label)
		c.w("%s:", body_label)
		bumped := c.addk(steps, 1)
		c.w("  store i64 %s, i64* @.vm.steps, align 8", bumped)
	}

	c.w("  call void @vm.guard()")

	switch inst.Name {

	case "PUSH":
		// push(): the whole Value_Holder is copied, sentinels included.
		ss := c.load_ss()
		pi := c.slot(ss, 0)
		c.w("  store i64 %d, i64* %s, align 8", inst.Operand.int64holder, pi)
		pf := c.slot(ss, 1)
		c.w("  store double %s, double* %s, align 8", llvmDouble(inst.Operand.float64holder), pf)
		pp := c.slot(ss, 2)
		if inst.Operand.pointer == "" {
			c.w("  store i8* null, i8** %s, align 8", pp)
		} else {
			c.w("  store i8* %s, i8** %s, align 8", c.cstr(inst.Operand.pointer), pp)
		}
		c.store_ss(c.addk(ss, 1))
		c.w("  br label %%%s", inst_label(next))

	case "ADDI":
		c.emit_binop_int("add", "Not enough values to add")
		c.w("  br label %%%s", inst_label(next))
	case "SUBI":
		c.emit_binop_int("sub", "Not enough values to subtract")
		c.w("  br label %%%s", inst_label(next))
	case "MULI":
		c.emit_binop_int("mul", "Not enough values to multiply")
		c.w("  br label %%%s", inst_label(next))
	case "DIVI":
		c.emit_binop_int("div", "Not enough values to divide")
		c.w("  br label %%%s", inst_label(next))
	case "EQI":
		c.emit_binop_int("eq", "Not enough values for equality")
		c.w("  br label %%%s", inst_label(next))

	case "ADDF":
		c.emit_binop_float("add", "Not enough values to add")
		c.w("  br label %%%s", inst_label(next))
	case "SUBF":
		c.emit_binop_float("sub", "Not enough values to subtract")
		c.w("  br label %%%s", inst_label(next))
	case "MULF":
		c.emit_binop_float("mul", "Not enough values to multiply")
		c.w("  br label %%%s", inst_label(next))
	case "DIVF":
		c.emit_binop_float("div", "Not enough values to divide")
		c.w("  br label %%%s", inst_label(next))
	case "EQF":
		c.emit_binop_float("eq", "Not enough values for equality")
		c.w("  br label %%%s", inst_label(next))

	case "JMP":
		target := inst.Operand.int64holder
		if target < 0 {
			c.w("  call void @vm.fatal(i8* $S[[Wrong Jump Instruction. Underflow]])")
			c.w("  unreachable")
		} else if target >= c.program_size {
			c.w("  call void @vm.fatal(i8* $S[[Wrong Jump Instruction. Overflow]])")
			c.w("  unreachable")
		} else {
			c.w("  br label %%%s", inst_label(target))
		}

	case "JMP_IF":
		target := inst.Operand.int64holder
		if target >= c.program_size {
			c.w("  call void @vm.fatal(i8* $S[[Wrong Jump_If Instruction. Overflow]])")
			c.w("  unreachable")
			break
		}
		ss := c.reg()
		c.w("  %s = call i64 @vm.need(i64 1, i8* $S[[Wrong Jump_If Instruction. Underflow]])", ss)
		top := c.addk(ss, -1)
		// jmp_if() tests `operand_type_check(top,"int64") && inst.Operand != 0`.
		// It never looks at the *value* on top of the stack - the second half of
		// the condition is the instruction's own operand. That is preserved here
		// so compiled programs branch exactly like interpreted ones.
		t := c.reg()
		c.w("  %s = call i32 @vm.typeof(i64 %s)", t, top)
		cond := c.reg()
		c.w("  %s = icmp eq i32 %s, 0", cond, t)
		c.store_ss(top)
		if target == 0 {
			c.w("  br label %%%s", inst_label(next))
		} else if target < 0 {
			taken := fmt.Sprintf("ip.%d.taken", ip)
			c.w("  br i1 %s, label %%%s, label %%%s", cond, taken, inst_label(next))
			c.w("%s:", taken)
			c.w("  call void @vm.oob(i64 %d)", target)
			c.w("  unreachable")
		} else {
			c.w("  br i1 %s, label %%%s, label %%%s", cond, inst_label(target), inst_label(next))
		}

	case "CALL":
		// call(): reset_operand_except() is a no-op, so only int64holder is
		// written. Whatever float/pointer the slot held before survives.
		ss := c.reg()
		c.w("  %s = load i64, i64* @.vm.ss, align 8", ss)
		full := c.reg()
		c.w("  %s = icmp sge i64 %s, %d", full, ss, STACK_CAPACITY)
		of := fmt.Sprintf("ip.%d.overflow", ip)
		body := fmt.Sprintf("ip.%d.push", ip)
		c.w("  br i1 %s, label %%%s, label %%%s", full, of, body)
		c.w("%s:", of)
		c.w("  call void @vm.fatal(i8* $S[[Stack Overflow]])")
		c.w("  unreachable")
		c.w("%s:", body)
		pi := c.slot(ss, 0)
		c.w("  store i64 %d, i64* %s, align 8", next, pi)
		c.store_ss(c.addk(ss, 1))
		c.branch_to(inst.Operand.int64holder)

	case "RET":
		ss := c.reg()
		c.w("  %s = call i64 @vm.need(i64 1, i8* $S[[Stack Underflow]])", ss)
		top := c.addk(ss, -1)
		dest := c.load_i(top)
		c.w("  store i64 %s, i64* @.vm.ip, align 8", dest)
		c.store_ss(top)
		c.w("  br label %%vm.dispatch")

	case "HALT":
		// halt() only sets vm_halt; inst_ptr is left alone and the loop ends.
		c.w("  br label %%vm.exit")

	case "NOP", "DEFINE", "INCLUDE", "IGNORE_HALT":
		c.w("  br label %%%s", inst_label(next))

	case "DROP":
		ss := c.reg()
		c.w("  %s = call i64 @vm.need(i64 1, i8* $S[[STACK UNDERFLOW]])", ss)
		c.store_ss(c.addk(ss, -1))
		c.w("  br label %%%s", inst_label(next))

	case "NOT":
		// not() only guards against stack_size < 0, so an empty stack indexes
		// STACK[-1] and panics in Go. Reproduce the panic instead of reading
		// out of bounds.
		ss := c.reg()
		c.w("  %s = call i64 @vm.need(i64 0, i8* $S[[Stack Underflow]])", ss)
		top := c.addk(ss, -1)
		neg := c.reg()
		c.w("  %s = icmp slt i64 %s, 0", neg, top)
		oob := fmt.Sprintf("ip.%d.oob", ip)
		body := fmt.Sprintf("ip.%d.not", ip)
		c.w("  br i1 %s, label %%%s, label %%%s", neg, oob, body)
		c.w("%s:", oob)
		c.w("  call void @vm.oob(i64 %s)", top)
		c.w("  unreachable")
		c.w("%s:", body)
		t := c.reg()
		c.w("  %s = call i32 @vm.typeof(i64 %s)", t, top)
		isint := c.reg()
		c.w("  %s = icmp eq i32 %s, 0", isint, t)
		pi := c.slot(top, 0)
		cur := c.reg()
		c.w("  %s = load i64, i64* %s, align 8", cur, pi)
		nonzero := c.reg()
		c.w("  %s = icmp ne i64 %s, 0", nonzero, cur)
		both := c.reg()
		c.w("  %s = and i1 %s, %s", both, isint, nonzero)
		res := c.reg()
		c.w("  %s = select i1 %s, i64 0, i64 1", res, both)
		c.w("  store i64 %s, i64* %s, align 8", res, pi)
		c.w("  br label %%%s", inst_label(next))

	case "DUP":
		n := inst.Operand.int64holder
		ss := c.reg()
		c.w("  %s = load i64, i64* @.vm.ss, align 8", ss)
		full := c.reg()
		c.w("  %s = icmp sge i64 %s, %d", full, ss, STACK_CAPACITY)
		of := fmt.Sprintf("ip.%d.overflow", ip)
		chk := fmt.Sprintf("ip.%d.chk", ip)
		c.w("  br i1 %s, label %%%s, label %%%s", full, of, chk)
		c.w("%s:", of)
		c.w("  call void @vm.fatal(i8* $S[[Stack Overflow]])")
		c.w("  unreachable")
		c.w("%s:", chk)
		depth := c.reg()
		c.w("  %s = sub nsw i64 %s, %d", depth, ss, n)
		under := c.reg()
		c.w("  %s = icmp sle i64 %s, 0", under, depth)
		uf := fmt.Sprintf("ip.%d.underflow", ip)
		body := fmt.Sprintf("ip.%d.dup", ip)
		c.w("  br i1 %s, label %%%s, label %%%s", under, uf, body)
		c.w("%s:", uf)
		c.w("  call void @vm.fatal(i8* $S[[Stack Underflow]])")
		c.w("  unreachable")
		c.w("%s:", body)
		src := c.addk(ss, -1-n)
		t := c.reg()
		c.w("  %s = call i32 @vm.typeof(i64 %s)", t, src)
		// dup() copies only the field matching the source type; a pointer copies
		// nothing at all, leaving the destination slot as it was.
		fl := fmt.Sprintf("ip.%d.dupf", ip)
		il := fmt.Sprintf("ip.%d.dupi", ip)
		jl := fmt.Sprintf("ip.%d.dupdone", ip)
		c.w("  switch i32 %s, label %%%s [", t, jl)
		c.w("    i32 0, label %%%s", il)
		c.w("    i32 1, label %%%s", fl)
		c.w("  ]")
		c.w("%s:", fl)
		sf := c.load_f(src)
		df := c.slot(ss, 1)
		c.w("  store double %s, double* %s, align 8", sf, df)
		c.w("  br label %%%s", jl)
		c.w("%s:", il)
		si := c.load_i(src)
		di := c.slot(ss, 0)
		c.w("  store i64 %s, i64* %s, align 8", si, di)
		c.w("  br label %%%s", jl)
		c.w("%s:", jl)
		c.store_ss(c.addk(ss, 1))
		c.w("  br label %%%s", inst_label(next))

	case "SWAP":
		n := inst.Operand.int64holder
		ss := c.reg()
		c.w("  %s = load i64, i64* @.vm.ss, align 8", ss)
		under := c.reg()
		c.w("  %s = icmp sge i64 %d, %s", under, n, ss)
		uf := fmt.Sprintf("ip.%d.underflow", ip)
		body := fmt.Sprintf("ip.%d.swap", ip)
		c.w("  br i1 %s, label %%%s, label %%%s", under, uf, body)
		c.w("%s:", uf)
		c.w("  call void @vm.fatal(i8* $S[[Stack Underflow]])")
		c.w("  unreachable")
		c.w("%s:", body)
		a := c.addk(ss, -1)
		b := c.addk(ss, -1-n)
		// A whole Value_Holder is exchanged, all three fields.
		pai, pbi := c.slot(a, 0), c.slot(b, 0)
		paf, pbf := c.slot(a, 1), c.slot(b, 1)
		pap, pbp := c.slot(a, 2), c.slot(b, 2)
		ai, bi := c.reg(), c.reg()
		c.w("  %s = load i64, i64* %s, align 8", ai, pai)
		c.w("  %s = load i64, i64* %s, align 8", bi, pbi)
		af, bf := c.reg(), c.reg()
		c.w("  %s = load double, double* %s, align 8", af, paf)
		c.w("  %s = load double, double* %s, align 8", bf, pbf)
		ap, bp := c.reg(), c.reg()
		c.w("  %s = load i8*, i8** %s, align 8", ap, pap)
		c.w("  %s = load i8*, i8** %s, align 8", bp, pbp)
		c.w("  store i64 %s, i64* %s, align 8", bi, pai)
		c.w("  store i64 %s, i64* %s, align 8", ai, pbi)
		c.w("  store double %s, double* %s, align 8", bf, paf)
		c.w("  store double %s, double* %s, align 8", af, pbf)
		c.w("  store i8* %s, i8** %s, align 8", bp, pap)
		c.w("  store i8* %s, i8** %s, align 8", ap, pbp)
		c.w("  br label %%%s", inst_label(next))

	case "PRINT":
		ss := c.reg()
		c.w("  %s = call i64 @vm.need(i64 1, i8* $S[[Not enough values on the stack to print]])", ss)
		top := c.addk(ss, -1)
		c.w("  call void @vm.emit_value(i64 %s)", top)
		c.store_ss(top)
		c.w("  br label %%%s", inst_label(next))

	case "PRINT_ASC":
		ss := c.reg()
		c.w("  %s = call i64 @vm.need(i64 1, i8* $S[[Not enough values on the stack to print]])", ss)
		top := c.addk(ss, -1)
		c.w("  call void @vm.print_asc(i64 %s)", top)
		c.store_ss(top)
		c.w("  br label %%%s", inst_label(next))

	default:
		exit_with_one(fmt.Sprintf("LLVM backend: unknown instruction `%s` at %d", inst.Name, ip))
	}
}

func describe_inst(inst Inst) string {
	switch inst.Name {
	case "PUSH":
		return fmt.Sprintf("PUSH {int64:%d float64:%v pointer:%q}",
			inst.Operand.int64holder, inst.Operand.float64holder, inst.Operand.pointer)
	case "JMP", "JMP_IF", "CALL", "DUP", "SWAP":
		return fmt.Sprintf("%s %d", inst.Name, inst.Operand.int64holder)
	default:
		return inst.Name
	}
}

// ---------------------------------------------------------------------------
// Driver
// ---------------------------------------------------------------------------

func llvm_output_path(input string, requested string) string {
	if requested != "" {
		return requested
	}
	base := strings.TrimSuffix(input, ".vasm")
	if base == input {
		base = strings.TrimSuffix(input, ".vm")
	}
	return base + ".ll"
}

// compile_program_to_llvm_ir lowers the program already loaded into vm to a
// textual LLVM IR module. `limit` mirrors the interpreter's -limit option; when
// `no_limit` is set the step counter is omitted entirely and the program runs
// until it halts.
func compile_program_to_llvm_ir(vm *VM, file_path string, out_path string, limit int, no_limit bool) {
	if vm.program_size == 0 {
		exit_with_one("Empty Program.. Cannot Compile to LLVM IR")
	}

	c := &llvm_compiler{
		str_names:    make(map[string]string),
		program_size: vm.program_size,
		limit:        limit,
		limited:      !no_limit,
	}
	if c.limited && limit <= 0 {
		// `counter < limit` is false from the start: nothing ever executes.
		c.limit = 0
	}

	for i := int64(0); i < vm.program_size; i++ {
		if vm.PROGRAM[i].Name == "RET" {
			c.needs_iptbl = true
		}
	}

	// ---- vm.main ----------------------------------------------------------
	c.w("; execute_program(): one basic block per VM instruction.")
	c.w("define internal void @vm.main() {")
	c.w("entry:")
	c.w("  br label %%%s", inst_label(0))

	for i := int64(0); i < vm.program_size; i++ {
		c.emit_instruction(i, vm.PROGRAM[i])
	}

	// Running off the end of the program is the interpreter's "Illegal
	// Instruction Access" path, and it is a real jump target (ip == program_size
	// after the last instruction), so it needs a block of its own.
	c.w("")
	c.w("; [%d] end of program", c.program_size)
	c.w("%s:", inst_label(c.program_size))
	if c.limited {
		steps := c.reg()
		c.w("  %s = load i64, i64* @.vm.steps, align 8", steps)
		ok := c.reg()
		c.w("  %s = icmp slt i64 %s, %d", ok, steps, c.limit)
		c.w("  br i1 %s, label %%ip.%d.body, label %%vm.exit", ok, c.program_size)
		c.w("ip.%d.body:", c.program_size)
	}
	c.w("  call void @vm.illegal()")
	c.w("  unreachable")

	if c.needs_iptbl {
		c.w("")
		c.w("; ret(): the return address is only known at runtime.")
		c.w("vm.dispatch:")
		ip := c.reg()
		c.w("  %s = load i64, i64* @.vm.ip, align 8", ip)
		neg := c.reg()
		c.w("  %s = icmp slt i64 %s, 0", neg, ip)
		big := c.reg()
		c.w("  %s = icmp sge i64 %s, %d", big, ip, PROGRAM_CAPACITY)
		bad := c.reg()
		c.w("  %s = or i1 %s, %s", bad, neg, big)
		c.w("  br i1 %s, label %%vm.dispatch.oob, label %%vm.dispatch.chk", bad)
		c.w("vm.dispatch.oob:")
		c.w("  call void @vm.oob(i64 %s)", ip)
		c.w("  unreachable")
		c.w("vm.dispatch.chk:")
		over := c.reg()
		c.w("  %s = icmp sgt i64 %s, %d", over, ip, c.program_size)
		c.w("  br i1 %s, label %%ip.%d, label %%vm.dispatch.go", over, c.program_size)
		c.w("vm.dispatch.go:")
		slot := c.reg()
		c.w("  %s = getelementptr inbounds [%d x i8*], [%d x i8*]* @.vm.iptable, i64 0, i64 %s",
			slot, c.program_size+1, c.program_size+1, ip)
		addr := c.reg()
		c.w("  %s = load i8*, i8** %s, align 8", addr, slot)
		targets := make([]string, 0, c.program_size+1)
		for i := int64(0); i <= c.program_size; i++ {
			targets = append(targets, "label %"+inst_label(i))
		}
		c.w("  indirectbr i8* %s, [ %s ]", addr, strings.Join(targets, ", "))
	}

	c.w("")
	c.w("vm.exit:")
	c.w("  ret void")
	c.w("}")

	c.w("")
	c.w("define i32 @main(i32 %%argc, i8** %%argv) {")
	c.w("entry:")
	c.w("  call void @vm.init()")
	c.w("  call void @vm.main()")
	c.w("  ret i32 0")
	c.w("}")

	// ---- assemble ---------------------------------------------------------
	replacer := strings.NewReplacer(
		"$BT", "`",
		"$CAP", strconv.Itoa(STACK_CAPACITY),
		"$PROGCAP", strconv.Itoa(PROGRAM_CAPACITY),
		"$MINFLOAT_G", "5e-324",
		"$MINFLOAT", llvmDouble(float64(MinFloat)),
		"$MININT", strconv.FormatInt(int64(MinInt), 10),
	)

	runtime := replacer.Replace(llvm_runtime_template)
	program := replacer.Replace(c.body.String())

	// $S[[..]] interning has to happen after substitution so the runtime's
	// sentinel spellings end up inside the literals.
	runtime = c.resolve_strings(runtime)
	program = c.resolve_strings(program)

	var module strings.Builder
	fmt.Fprintf(&module, "; LLVM IR generated by vm-go from %s\n", file_path)
	fmt.Fprintf(&module, "; %d instruction(s)", vm.program_size)
	if c.limited {
		fmt.Fprintf(&module, ", execution limit %d step(s)", c.limit)
	} else {
		fmt.Fprintf(&module, ", no execution limit")
	}
	module.WriteString("\n")
	fmt.Fprintf(&module, "; Build with: clang %s -o %s\n",
		llvm_output_path(file_path, out_path),
		strings.TrimSuffix(llvm_output_path(file_path, out_path), ".ll"))
	fmt.Fprintf(&module, "source_filename = \"%s\"\n", file_path)
	module.WriteString(runtime)
	module.WriteString("\n; --- interned string literals ---------------------------------------------\n")
	module.WriteString(c.string_globals())

	if c.needs_iptbl {
		entries := make([]string, 0, c.program_size+1)
		for i := int64(0); i <= c.program_size; i++ {
			entries = append(entries, fmt.Sprintf("i8* blockaddress(@vm.main, %%%s)", inst_label(i)))
		}
		fmt.Fprintf(&module, "\n@.vm.iptable = internal unnamed_addr constant [%d x i8*] [%s]\n",
			c.program_size+1, strings.Join(entries, ", "))
	}

	module.WriteString("\n")
	module.WriteString(program)

	target := llvm_output_path(file_path, out_path)
	if target == "-" {
		os.Stdout.WriteString(module.String())
		return
	}

	fmt.Printf("Compiling '%s' to LLVM IR '%s'\n", file_path, target)
	if err := os.WriteFile(target, []byte(module.String()), 0644); err != nil {
		fmt.Println(err)
		exit_with_one("Cannot write LLVM IR file")
	}
	fmt.Println("LLVM IR Written To:", target)
	fmt.Printf("Build it with: clang %s -o %s\n", target, strings.TrimSuffix(target, ".ll"))
	fmt.Println()
}
