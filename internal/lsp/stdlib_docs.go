package lsp

import (
	"github.com/d5/tengo/v2/parser"
	"github.com/d5/tengo/v2/stdlib"
)

type stdlibEntry struct {
	Sig string // displayed in code block
	Doc string // plain-text description (may be empty)
}

// stdlibDocs is the static documentation index for all built-in Tengo stdlib
// modules except enum (which is documented via its embedded Tengo source).
var stdlibDocs = map[string]map[string]stdlibEntry{
	"fmt": {
		"print":   {Sig: "print(args...)", Doc: "Prints args to stdout."},
		"println": {Sig: "println(args...)", Doc: "Prints args to stdout followed by a newline."},
		"printf":  {Sig: "printf(format, args...)", Doc: "Formats and prints to stdout."},
		"sprintf": {Sig: "sprintf(format, args...) string", Doc: "Returns a formatted string."},
	},

	"math": {
		// constants
		"e":                      {Sig: "e float", Doc: "Euler's number (2.71828…)."},
		"pi":                     {Sig: "pi float", Doc: "Pi (3.14159…)."},
		"phi":                    {Sig: "phi float", Doc: "Golden ratio (1.61803…)."},
		"sqrt2":                  {Sig: "sqrt2 float", Doc: "Square root of 2."},
		"sqrtE":                  {Sig: "sqrtE float", Doc: "Square root of e."},
		"sqrtPi":                 {Sig: "sqrtPi float", Doc: "Square root of Pi."},
		"sqrtPhi":                {Sig: "sqrtPhi float", Doc: "Square root of Phi."},
		"ln2":                    {Sig: "ln2 float", Doc: "Natural log of 2."},
		"log2E":                  {Sig: "log2E float", Doc: "Log base-2 of e."},
		"ln10":                   {Sig: "ln10 float", Doc: "Natural log of 10."},
		"log10E":                 {Sig: "log10E float", Doc: "Log base-10 of e."},
		"maxFloat32":             {Sig: "maxFloat32 float"},
		"smallestNonzeroFloat32": {Sig: "smallestNonzeroFloat32 float"},
		"maxFloat64":             {Sig: "maxFloat64 float"},
		"smallestNonzeroFloat64": {Sig: "smallestNonzeroFloat64 float"},
		"maxInt":                 {Sig: "maxInt int"},
		"minInt":                 {Sig: "minInt int"},
		"maxInt8":                {Sig: "maxInt8 int"},
		"minInt8":                {Sig: "minInt8 int"},
		"maxInt16":               {Sig: "maxInt16 int"},
		"minInt16":               {Sig: "minInt16 int"},
		"maxInt32":               {Sig: "maxInt32 int"},
		"minInt32":               {Sig: "minInt32 int"},
		"maxInt64":               {Sig: "maxInt64 int"},
		"minInt64":               {Sig: "minInt64 int"},
		// single-arg functions
		"abs":   {Sig: "abs(x float) float", Doc: "Absolute value."},
		"acos":  {Sig: "acos(x float) float", Doc: "Arccosine in radians."},
		"acosh": {Sig: "acosh(x float) float", Doc: "Inverse hyperbolic cosine."},
		"asin":  {Sig: "asin(x float) float", Doc: "Arcsine in radians."},
		"asinh": {Sig: "asinh(x float) float", Doc: "Inverse hyperbolic sine."},
		"atan":  {Sig: "atan(x float) float", Doc: "Arctangent in radians."},
		"atanh": {Sig: "atanh(x float) float", Doc: "Inverse hyperbolic tangent."},
		"cbrt":  {Sig: "cbrt(x float) float", Doc: "Cube root."},
		"ceil":  {Sig: "ceil(x float) float", Doc: "Rounds up to nearest integer."},
		"cos":   {Sig: "cos(x float) float", Doc: "Cosine of x radians."},
		"cosh":  {Sig: "cosh(x float) float", Doc: "Hyperbolic cosine."},
		"erf":   {Sig: "erf(x float) float", Doc: "Error function."},
		"erfc":  {Sig: "erfc(x float) float", Doc: "Complementary error function."},
		"exp":   {Sig: "exp(x float) float", Doc: "e**x."},
		"exp2":  {Sig: "exp2(x float) float", Doc: "2**x."},
		"expm1": {Sig: "expm1(x float) float", Doc: "e**x - 1 (accurate for small x)."},
		"floor": {Sig: "floor(x float) float", Doc: "Rounds down to nearest integer."},
		"gamma": {Sig: "gamma(x float) float", Doc: "Gamma function."},
		"j0":    {Sig: "j0(x float) float", Doc: "Bessel function of the first kind, order 0."},
		"j1":    {Sig: "j1(x float) float", Doc: "Bessel function of the first kind, order 1."},
		"log":   {Sig: "log(x float) float", Doc: "Natural logarithm."},
		"log10": {Sig: "log10(x float) float", Doc: "Decimal logarithm."},
		"log1p": {Sig: "log1p(x float) float", Doc: "Natural log of 1+x (accurate for small x)."},
		"log2":  {Sig: "log2(x float) float", Doc: "Binary logarithm."},
		"logb":  {Sig: "logb(x float) float", Doc: "Binary exponent of x."},
		"nan":   {Sig: "nan() float", Doc: "IEEE 754 NaN."},
		"sin":   {Sig: "sin(x float) float", Doc: "Sine of x radians."},
		"sinh":  {Sig: "sinh(x float) float", Doc: "Hyperbolic sine."},
		"sqrt":  {Sig: "sqrt(x float) float", Doc: "Square root."},
		"tan":   {Sig: "tan(x float) float", Doc: "Tangent of x radians."},
		"tanh":  {Sig: "tanh(x float) float", Doc: "Hyperbolic tangent."},
		"trunc": {Sig: "trunc(x float) float", Doc: "Truncates to integer part."},
		"y0":    {Sig: "y0(x float) float", Doc: "Bessel function of the second kind, order 0."},
		"y1":    {Sig: "y1(x float) float", Doc: "Bessel function of the second kind, order 1."},
		// two-arg functions
		"atan2":     {Sig: "atan2(y, x float) float", Doc: "Arctangent of y/x."},
		"copysign":  {Sig: "copysign(x, y float) float", Doc: "x with the sign of y."},
		"dim":       {Sig: "dim(x, y float) float", Doc: "max(x-y, 0)."},
		"hypot":     {Sig: "hypot(p, q float) float", Doc: "sqrt(p*p + q*q)."},
		"max":       {Sig: "max(x, y float) float", Doc: "Larger of x or y."},
		"min":       {Sig: "min(x, y float) float", Doc: "Smaller of x or y."},
		"mod":       {Sig: "mod(x, y float) float", Doc: "Floating-point remainder of x/y."},
		"nextafter": {Sig: "nextafter(x, y float) float", Doc: "Next float after x toward y."},
		"pow":       {Sig: "pow(x, y float) float", Doc: "x**y."},
		"remainder": {Sig: "remainder(x, y float) float", Doc: "IEEE 754 remainder of x/y."},
		// special
		"ilogb":    {Sig: "ilogb(x float) int", Doc: "Binary exponent of x as integer."},
		"inf":      {Sig: "inf(sign int) float", Doc: "Positive/negative infinity (sign >= 0 → +Inf)."},
		"is_inf":   {Sig: "is_inf(f float, sign int) bool", Doc: "Reports whether f is infinity (sign: 1=+Inf, -1=-Inf, 0=either)."},
		"is_nan":   {Sig: "is_nan(f float) bool", Doc: "Reports whether f is NaN."},
		"jn":       {Sig: "jn(n int, x float) float", Doc: "Bessel function of the first kind, order n."},
		"ldexp":    {Sig: "ldexp(frac float, exp int) float", Doc: "frac × 2**exp."},
		"pow10":    {Sig: "pow10(n int) float", Doc: "10**n."},
		"signbit":  {Sig: "signbit(x float) bool", Doc: "Reports whether x is negative or negative zero."},
		"yn":       {Sig: "yn(n int, x float) float", Doc: "Bessel function of the second kind, order n."},
	},

	"os": {
		"platform":         {Sig: "platform string", Doc: "Operating system (e.g. \"linux\", \"darwin\", \"windows\")."},
		"arch":             {Sig: "arch string", Doc: "CPU architecture (e.g. \"amd64\", \"arm64\")."},
		"args":             {Sig: "args() []string", Doc: "Command-line arguments."},
		"exit":             {Sig: "exit(code int)", Doc: "Exits the process with the given status code."},
		"getenv":           {Sig: "getenv(name string) string", Doc: "Returns the value of an environment variable."},
		"setenv":           {Sig: "setenv(name, value string) error", Doc: "Sets an environment variable."},
		"unsetenv":         {Sig: "unsetenv(name string) error", Doc: "Unsets an environment variable."},
		"expand_env":       {Sig: "expand_env(s string) string", Doc: "Replaces ${VAR} and $VAR in s with their environment values."},
		"environ":          {Sig: "environ() []string", Doc: "Returns all environment variables as key=value strings."},
		"clearenv":         {Sig: "clearenv()", Doc: "Deletes all environment variables."},
		"lookup_env":       {Sig: "lookup_env(name string) (string, bool)", Doc: "Returns env var value and whether it is set."},
		"getwd":            {Sig: "getwd() string", Doc: "Returns the current working directory."},
		"chdir":            {Sig: "chdir(path string) error", Doc: "Changes the current working directory."},
		"hostname":         {Sig: "hostname() string", Doc: "Returns the host name."},
		"getuid":           {Sig: "getuid() int", Doc: "Returns the numeric user ID of the caller."},
		"getgid":           {Sig: "getgid() int", Doc: "Returns the numeric group ID of the caller."},
		"geteuid":          {Sig: "geteuid() int", Doc: "Returns the numeric effective user ID."},
		"getegid":          {Sig: "getegid() int", Doc: "Returns the numeric effective group ID."},
		"getgroups":        {Sig: "getgroups() []int", Doc: "Returns supplementary group IDs."},
		"getpid":           {Sig: "getpid() int", Doc: "Returns the process ID."},
		"getppid":          {Sig: "getppid() int", Doc: "Returns the parent process ID."},
		"getpagesize":      {Sig: "getpagesize() int", Doc: "Returns the memory page size."},
		"read_file":        {Sig: "read_file(filename string) bytes", Doc: "Reads the named file and returns the contents."},
		"write_file":       {Sig: "write_file(filename string, data bytes, perm int) error", Doc: "Writes data to the named file, creating it if needed."},
		"open":             {Sig: "open(name string) file", Doc: "Opens a file for reading."},
		"open_file":        {Sig: "open_file(name string, flag, perm int) file", Doc: "Opens a file with the given flags and permissions."},
		"create":           {Sig: "create(name string) file", Doc: "Creates or truncates the named file."},
		"stat":             {Sig: "stat(name string) info", Doc: "Returns file info for the named file."},
		"link":             {Sig: "link(oldname, newname string) error", Doc: "Creates a hard link."},
		"symlink":          {Sig: "symlink(oldname, newname string) error", Doc: "Creates a symbolic link."},
		"readlink":         {Sig: "readlink(name string) string", Doc: "Returns the destination of a symbolic link."},
		"mkdir":            {Sig: "mkdir(name string, perm int) error", Doc: "Creates a directory."},
		"mkdir_all":        {Sig: "mkdir_all(path string, perm int) error", Doc: "Creates a directory and all parents."},
		"remove":           {Sig: "remove(name string) error", Doc: "Removes a file or empty directory."},
		"remove_all":       {Sig: "remove_all(path string) error", Doc: "Removes path and all children."},
		"rename":           {Sig: "rename(oldpath, newpath string) error", Doc: "Renames a file or directory."},
		"chmod":            {Sig: "chmod(name string, mode int) error", Doc: "Changes the file mode."},
		"chown":            {Sig: "chown(name string, uid, gid int) error", Doc: "Changes the file owner."},
		"lchown":           {Sig: "lchown(name string, uid, gid int) error", Doc: "Changes the owner of a symbolic link."},
		"truncate":         {Sig: "truncate(name string, size int) error", Doc: "Truncates a file to the given size."},
		"temp_dir":         {Sig: "temp_dir() string", Doc: "Returns the default directory for temporary files."},
		"find_process":     {Sig: "find_process(pid int) process", Doc: "Finds the process with the given PID."},
		"start_process":    {Sig: "start_process(name string, argv []string, attr map) process", Doc: "Starts a new process."},
		"exec":             {Sig: "exec(name string, args...string) cmd", Doc: "Prepares a command for execution."},
		"exec_look_path":   {Sig: "exec_look_path(file string) string", Doc: "Searches for an executable in PATH."},
		"pipe":             {Sig: "pipe() (read, write file)", Doc: "Creates a synchronous in-memory pipe."},
		"stdin":            {Sig: "stdin file", Doc: "Standard input."},
		"stdout":           {Sig: "stdout file", Doc: "Standard output."},
		"stderr":           {Sig: "stderr file", Doc: "Standard error."},
		"dev_null":         {Sig: "dev_null string", Doc: "Path of the null device (e.g. \"/dev/null\")."},
		"path_separator":   {Sig: "path_separator string", Doc: "OS-specific path separator (e.g. \"/\")."},
		"path_list_separator": {Sig: "path_list_separator string", Doc: "OS-specific path list separator (e.g. \":\")."},
		// file mode constants
		"o_rdonly": {Sig: "o_rdonly int", Doc: "Open file read-only."},
		"o_wronly": {Sig: "o_wronly int", Doc: "Open file write-only."},
		"o_rdwr":   {Sig: "o_rdwr int", Doc: "Open file read-write."},
		"o_append": {Sig: "o_append int", Doc: "Append data to the file when writing."},
		"o_create": {Sig: "o_create int", Doc: "Create the file if it doesn't exist."},
		"o_excl":   {Sig: "o_excl int", Doc: "File must not exist when opening."},
		"o_sync":   {Sig: "o_sync int", Doc: "Open for synchronous I/O."},
		"o_trunc":  {Sig: "o_trunc int", Doc: "Truncate file when opening."},
		"seek_set": {Sig: "seek_set int", Doc: "Seek relative to the start of the file."},
		"seek_cur": {Sig: "seek_cur int", Doc: "Seek relative to the current position."},
		"seek_end": {Sig: "seek_end int", Doc: "Seek relative to the end of the file."},
	},

	"text": {
		"re_match":      {Sig: "re_match(pattern, text string) bool", Doc: "Reports whether text contains a match for pattern."},
		"re_find":       {Sig: "re_find(pattern, text string, n int) [][]string", Doc: "Returns up to n matches of pattern in text (-1 = all)."},
		"re_replace":    {Sig: "re_replace(pattern, text, repl string) string", Doc: "Replaces matches of pattern in text with repl."},
		"re_split":      {Sig: "re_split(pattern, text string, n int) []string", Doc: "Splits text by pattern."},
		"re_compile":    {Sig: "re_compile(pattern string) regexp", Doc: "Compiles a regular expression."},
		"compare":       {Sig: "compare(a, b string) int", Doc: "Lexicographic comparison: -1, 0, or 1."},
		"contains":      {Sig: "contains(s, substr string) bool", Doc: "Reports whether substr is in s."},
		"contains_any":  {Sig: "contains_any(s, chars string) bool", Doc: "Reports whether any Unicode code point in chars is in s."},
		"count":         {Sig: "count(s, substr string) int", Doc: "Counts non-overlapping occurrences of substr in s."},
		"equal_fold":    {Sig: "equal_fold(s, t string) bool", Doc: "Case-insensitive equality check."},
		"fields":        {Sig: "fields(s string) []string", Doc: "Splits s around each whitespace run."},
		"has_prefix":    {Sig: "has_prefix(s, prefix string) bool", Doc: "Reports whether s begins with prefix."},
		"has_suffix":    {Sig: "has_suffix(s, suffix string) bool", Doc: "Reports whether s ends with suffix."},
		"index":         {Sig: "index(s, substr string) int", Doc: "First index of substr in s, or -1."},
		"index_any":     {Sig: "index_any(s, chars string) int", Doc: "First index of any char in s, or -1."},
		"join":          {Sig: "join(elems []string, sep string) string", Doc: "Concatenates elems with sep between them."},
		"last_index":    {Sig: "last_index(s, substr string) int", Doc: "Last index of substr in s, or -1."},
		"last_index_any": {Sig: "last_index_any(s, chars string) int", Doc: "Last index of any Unicode code point from chars in s."},
		"pad_left":      {Sig: "pad_left(s string, pad_len int, pad_char string) string", Doc: "Pads s on the left to pad_len."},
		"pad_right":     {Sig: "pad_right(s string, pad_len int, pad_char string) string", Doc: "Pads s on the right to pad_len."},
		"repeat":        {Sig: "repeat(s string, count int) string", Doc: "Returns s repeated count times."},
		"replace":       {Sig: "replace(s, old, new string, n int) string", Doc: "Replaces up to n occurrences of old with new in s."},
		"split":         {Sig: "split(s, sep string) []string", Doc: "Splits s into all substrings separated by sep."},
		"split_n":       {Sig: "split_n(s, sep string, n int) []string", Doc: "Splits s into at most n substrings."},
		"split_after":   {Sig: "split_after(s, sep string) []string", Doc: "Splits s after each instance of sep."},
		"split_after_n": {Sig: "split_after_n(s, sep string, n int) []string", Doc: "Splits s after sep, at most n substrings."},
		"substr":        {Sig: "substr(s string, low, high int) string", Doc: "Returns s[low:high]."},
		"title":         {Sig: "title(s string) string", Doc: "Returns s with the first letter of each word capitalised."},
		"to_lower":      {Sig: "to_lower(s string) string", Doc: "Returns s in lower case."},
		"to_title":      {Sig: "to_title(s string) string", Doc: "Returns s with all characters mapped to title case."},
		"to_upper":      {Sig: "to_upper(s string) string", Doc: "Returns s in upper case."},
		"trim":          {Sig: "trim(s, cutset string) string", Doc: "Removes leading and trailing characters in cutset."},
		"trim_left":     {Sig: "trim_left(s, cutset string) string", Doc: "Removes leading characters in cutset."},
		"trim_right":    {Sig: "trim_right(s, cutset string) string", Doc: "Removes trailing characters in cutset."},
		"trim_prefix":   {Sig: "trim_prefix(s, prefix string) string", Doc: "Removes prefix from s."},
		"trim_suffix":   {Sig: "trim_suffix(s, suffix string) string", Doc: "Removes suffix from s."},
		"trim_space":    {Sig: "trim_space(s string) string", Doc: "Removes leading and trailing whitespace."},
		"atoi":          {Sig: "atoi(s string) int", Doc: "Parses s as a base-10 integer."},
		"itoa":          {Sig: "itoa(i int) string", Doc: "Returns the string representation of i."},
		"format_bool":   {Sig: "format_bool(b bool) string", Doc: "Returns \"true\" or \"false\"."},
		"parse_bool":    {Sig: "parse_bool(s string) bool", Doc: "Parses a boolean string (\"1\", \"t\", \"true\", \"0\", \"f\", \"false\")."},
		"format_float":  {Sig: "format_float(f float, fmt string, prec, bitSize int) string", Doc: "Formats a float."},
		"parse_float":   {Sig: "parse_float(s string, bitSize int) float", Doc: "Parses a floating-point number."},
		"parse_int":     {Sig: "parse_int(s string, base, bitSize int) int", Doc: "Parses an integer in the given base."},
		"format_int":    {Sig: "format_int(i int, base int) string", Doc: "Returns the string representation of i in the given base."},
		"quote":         {Sig: "quote(s string) string", Doc: "Returns a Go-syntax double-quoted string literal."},
		"unquote":       {Sig: "unquote(s string) string", Doc: "Interprets s as a Go-syntax quoted string."},
	},

	"times": {
		// format constants
		"format_ansic":        {Sig: "format_ansic string", Doc: "\"Mon Jan _2 15:04:05 2006\""},
		"format_unix_date":    {Sig: "format_unix_date string", Doc: "\"Mon Jan _2 15:04:05 MST 2006\""},
		"format_ruby_date":    {Sig: "format_ruby_date string", Doc: "\"Mon Jan 02 15:04:05 -0700 2006\""},
		"format_rfc822":       {Sig: "format_rfc822 string", Doc: "\"02 Jan 06 15:04 MST\""},
		"format_rfc822z":      {Sig: "format_rfc822z string", Doc: "\"02 Jan 06 15:04 -0700\""},
		"format_rfc850":       {Sig: "format_rfc850 string", Doc: "\"Monday, 02-Jan-06 15:04:05 MST\""},
		"format_rfc1123":      {Sig: "format_rfc1123 string", Doc: "\"Mon, 02 Jan 2006 15:04:05 MST\""},
		"format_rfc1123z":     {Sig: "format_rfc1123z string", Doc: "\"Mon, 02 Jan 2006 15:04:05 -0700\""},
		"format_rfc3339":      {Sig: "format_rfc3339 string", Doc: "\"2006-01-02T15:04:05Z07:00\""},
		"format_rfc3339_nano": {Sig: "format_rfc3339_nano string", Doc: "\"2006-01-02T15:04:05.999999999Z07:00\""},
		"format_kitchen":      {Sig: "format_kitchen string", Doc: "\"3:04PM\""},
		"format_stamp":        {Sig: "format_stamp string", Doc: "\"Jan _2 15:04:05\""},
		"format_stamp_milli":  {Sig: "format_stamp_milli string", Doc: "\"Jan _2 15:04:05.000\""},
		"format_stamp_micro":  {Sig: "format_stamp_micro string", Doc: "\"Jan _2 15:04:05.000000\""},
		"format_stamp_nano":   {Sig: "format_stamp_nano string", Doc: "\"Jan _2 15:04:05.000000000\""},
		// duration constants
		"nanosecond":  {Sig: "nanosecond int"},
		"microsecond": {Sig: "microsecond int"},
		"millisecond": {Sig: "millisecond int"},
		"second":      {Sig: "second int"},
		"minute":      {Sig: "minute int"},
		"hour":        {Sig: "hour int"},
		// month constants
		"january": {Sig: "january int"}, "february": {Sig: "february int"},
		"march": {Sig: "march int"}, "april": {Sig: "april int"},
		"may": {Sig: "may int"}, "june": {Sig: "june int"},
		"july": {Sig: "july int"}, "august": {Sig: "august int"},
		"september": {Sig: "september int"}, "october": {Sig: "october int"},
		"november": {Sig: "november int"}, "december": {Sig: "december int"},
		// functions
		"now":                 {Sig: "now() time", Doc: "Returns the current local time."},
		"since":               {Sig: "since(t time) int", Doc: "Returns nanoseconds elapsed since t."},
		"until":               {Sig: "until(t time) int", Doc: "Returns nanoseconds until t."},
		"sleep":               {Sig: "sleep(duration int)", Doc: "Pauses for the given duration in nanoseconds."},
		"parse":               {Sig: "parse(layout, value string) time", Doc: "Parses a formatted time string using layout."},
		"parse_duration":      {Sig: "parse_duration(s string) int", Doc: "Parses a duration string (e.g. \"1h30m\") into nanoseconds."},
		"unix":                {Sig: "unix(sec, nsec int) time", Doc: "Returns the local time for the Unix epoch offset."},
		"date":                {Sig: "date(year, month, day, hour, min, sec, nsec int, loc string) time", Doc: "Returns the time at the given date and time."},
		"add":                 {Sig: "add(t time, duration int) time", Doc: "Adds a nanosecond duration to t."},
		"add_date":            {Sig: "add_date(t time, years, months, days int) time", Doc: "Adds years, months, days to t."},
		"sub":                 {Sig: "sub(t, u time) int", Doc: "Returns t-u in nanoseconds."},
		"after":               {Sig: "after(t, u time) bool", Doc: "Reports whether t is after u."},
		"before":              {Sig: "before(t, u time) bool", Doc: "Reports whether t is before u."},
		"time_year":           {Sig: "time_year(t time) int", Doc: "Returns the year."},
		"time_month":          {Sig: "time_month(t time) int", Doc: "Returns the month (1–12)."},
		"time_day":            {Sig: "time_day(t time) int", Doc: "Returns the day of the month."},
		"time_hour":           {Sig: "time_hour(t time) int", Doc: "Returns the hour (0–23)."},
		"time_minute":         {Sig: "time_minute(t time) int", Doc: "Returns the minute (0–59)."},
		"time_second":         {Sig: "time_second(t time) int", Doc: "Returns the second (0–59)."},
		"time_nanosecond":     {Sig: "time_nanosecond(t time) int", Doc: "Returns the nanosecond."},
		"time_unix":           {Sig: "time_unix(t time) int", Doc: "Returns Unix time in seconds."},
		"time_unix_nano":      {Sig: "time_unix_nano(t time) int", Doc: "Returns Unix time in nanoseconds."},
		"time_format":         {Sig: "time_format(t time, layout string) string", Doc: "Formats t using layout."},
		"time_location":       {Sig: "time_location(t time) string", Doc: "Returns the time zone name."},
		"time_weekday":        {Sig: "time_weekday(t time) int", Doc: "Returns the day of the week (0=Sunday)."},
		"time_string":         {Sig: "time_string(t time) string", Doc: "Returns a human-readable representation of t."},
		"time":                {Sig: "time(t time) time", Doc: "Creates a time value."},
		"is_zero":             {Sig: "is_zero(t time) bool", Doc: "Reports whether t is the zero time."},
		"to_local":            {Sig: "to_local(t time) time", Doc: "Returns t in local time zone."},
		"to_utc":              {Sig: "to_utc(t time) time", Doc: "Returns t in UTC."},
		"in_location":         {Sig: "in_location(t time, loc string) time", Doc: "Returns t in the named location."},
		"duration_hours":      {Sig: "duration_hours(d int) float", Doc: "Duration as floating-point hours."},
		"duration_minutes":    {Sig: "duration_minutes(d int) float", Doc: "Duration as floating-point minutes."},
		"duration_seconds":    {Sig: "duration_seconds(d int) float", Doc: "Duration as floating-point seconds."},
		"duration_nanoseconds": {Sig: "duration_nanoseconds(d int) int", Doc: "Duration as integer nanoseconds."},
		"duration_string":     {Sig: "duration_string(d int) string", Doc: "Returns a string like \"72h3m0.5s\"."},
		"month_string":        {Sig: "month_string(m int) string", Doc: "Returns the English name of month m."},
		"truncate":            {Sig: "truncate(t time, d int) time", Doc: "Rounds t down to a multiple of duration d."},
		"round":               {Sig: "round(t time, d int) time", Doc: "Rounds t to the nearest multiple of duration d."},
	},

	"rand": {
		"int":        {Sig: "int() int", Doc: "Returns a random non-negative int."},
		"float":      {Sig: "float() float", Doc: "Returns a random float in [0.0, 1.0)."},
		"intn":       {Sig: "intn(n int) int", Doc: "Returns a random int in [0, n)."},
		"exp_float":  {Sig: "exp_float() float", Doc: "Returns an exponentially distributed float."},
		"norm_float": {Sig: "norm_float() float", Doc: "Returns a normally distributed float with mean 0, stddev 1."},
		"perm":       {Sig: "perm(n int) []int", Doc: "Returns a random permutation of [0, n)."},
		"seed":       {Sig: "seed(seed int)", Doc: "Sets the global random seed."},
		"read":       {Sig: "read(b bytes) int", Doc: "Fills b with random bytes."},
		"rand":       {Sig: "rand(seed int) rand", Doc: "Returns a new rand source seeded with seed."},
	},

	"json": {
		"decode":      {Sig: "decode(s string) any", Doc: "Unmarshals a JSON string into a Tengo value."},
		"encode":      {Sig: "encode(v any) string", Doc: "Marshals a Tengo value to a JSON string."},
		"indent":      {Sig: "indent(s, prefix, indent string) string", Doc: "Indents a JSON string."},
		"html_escape": {Sig: "html_escape(s string) string", Doc: "Applies HTML escaping to a JSON string."},
	},

	"hex": {
		"encode": {Sig: "encode(b bytes) string", Doc: "Returns the hex encoding of b."},
		"decode": {Sig: "decode(s string) bytes", Doc: "Decodes a hex string."},
	},

	"base64": {
		"encode":         {Sig: "encode(b bytes) string", Doc: "Standard base64 encoding with padding."},
		"decode":         {Sig: "decode(s string) bytes", Doc: "Decodes standard base64."},
		"raw_encode":     {Sig: "raw_encode(b bytes) string", Doc: "Standard base64 encoding without padding."},
		"raw_decode":     {Sig: "raw_decode(s string) bytes", Doc: "Decodes standard base64 without padding."},
		"url_encode":     {Sig: "url_encode(b bytes) string", Doc: "URL-safe base64 encoding with padding."},
		"url_decode":     {Sig: "url_decode(s string) bytes", Doc: "Decodes URL-safe base64."},
		"raw_url_encode": {Sig: "raw_url_encode(b bytes) string", Doc: "URL-safe base64 encoding without padding."},
		"raw_url_decode": {Sig: "raw_url_decode(s string) bytes", Doc: "Decodes URL-safe base64 without padding."},
	},
}

// hoverStdlib returns hover markdown for a stdlib module member.
// enum is handled by parsing its embedded Tengo source (which contains comments).
func hoverStdlib(mod, name string) string {
	if mod == "enum" {
		return hoverEnumMember(name)
	}
	m, ok := stdlibDocs[mod]
	if !ok {
		return "```tengo\n" + name + "\n```"
	}
	e, ok := m[name]
	if !ok {
		return "```tengo\n" + name + "\n```"
	}
	if e.Doc != "" {
		return "```tengo\n" + e.Sig + "\n```\n" + e.Doc
	}
	return "```tengo\n" + e.Sig + "\n```"
}

// hoverEnumMember parses the embedded Tengo source for the enum stdlib module
// and extracts the function signature and leading comment for name.
func hoverEnumMember(name string) string {
	src := stdlib.SourceModules["enum"]
	file, srcFile, _ := parseDoc(src)
	if file == nil {
		return "```tengo\n" + name + "\n```"
	}
	for _, stmt := range file.Stmts {
		exp, ok := stmt.(*parser.ExportStmt)
		if !ok {
			continue
		}
		mapLit, ok := exp.Result.(*parser.MapLit)
		if !ok {
			continue
		}
		for _, elem := range mapLit.Elements {
			if elem.Key != name {
				continue
			}
			fn, ok := elem.Value.(*parser.FuncLit)
			if !ok {
				return "```tengo\n" + name + "\n```"
			}
			sig := funcSignature(name, fn)
			keyLine := srcFile.Position(elem.KeyPos).Line - 1
			cmt := leadingComment(src, keyLine)
			if cmt != "" {
				return "```tengo\n" + sig + "\n```\n" + cmt
			}
			return "```tengo\n" + sig + "\n```"
		}
	}
	return "```tengo\n" + name + "\n```"
}
