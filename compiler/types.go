package compiler

type Type int

const (
	TInvalid Type = iota
	TInt
	TBool
	TStr
	TFloat
	TAny
	TSliceInt
	TSliceFloat
	TSliceBool
	TSliceStr
)

func (t Type) String() string {
	switch t {
	case TInt:
		return "i32"
	case TBool:
		return "bool"
	case TStr:
		return "string"
	case TFloat:
		return "f64"
	case TAny:
		return "any"
	case TSliceInt:
		return "[]i32"
	case TSliceFloat:
		return "[]f64"
	case TSliceBool:
		return "[]bool"
	case TSliceStr:
		return "[]string"
	}
	return "invalid"
}

func typeFromName(name string) Type {
	switch name {
	case "i32":
		return TInt
	case "bool":
		return TBool
	case "string":
		return TStr
	case "f64":
		return TFloat
	}
	if len(name) >= 2 && name[0] == '[' && name[1] == ']' {
		switch typeFromName(name[2:]) {
		case TInt:
			return TSliceInt
		case TFloat:
			return TSliceFloat
		case TBool:
			return TSliceBool
		case TStr:
			return TSliceStr
		}
	}
	return TInvalid
}

func isSlice(t Type) bool {
	return sliceElem(t) != TInvalid
}

func sliceElem(t Type) Type {
	switch t {
	case TSliceInt:
		return TInt
	case TSliceFloat:
		return TFloat
	case TSliceBool:
		return TBool
	case TSliceStr:
		return TStr
	}
	return TInvalid
}

func sliceOf(elem Type) Type {
	switch elem {
	case TInt:
		return TSliceInt
	case TFloat:
		return TSliceFloat
	case TBool:
		return TSliceBool
	case TStr:
		return TSliceStr
	}
	return TInvalid
}

func irType(t Type) string {
	switch t {
	case TInt:
		return "i64"
	case TBool:
		return "i1"
	case TStr:
		return "i8*"
	case TFloat:
		return "double"
	case TSliceInt, TSliceFloat, TSliceBool, TSliceStr:
		return "i8*"
	}
	return "i64"
}

func abiType(t Type) string {
	switch t {
	case TInt:
		return "i64"
	case TBool:
		return "i32"
	case TStr:
		return "i8*"
	case TFloat:
		return "double"
	case TSliceInt, TSliceFloat, TSliceBool, TSliceStr:
		return "i8*"
	}
	return "i64"
}

type builtin struct {
	ret   Type
	args  []Type
	cname string
	print string
}

func bi(ret Type, cname string, args ...Type) builtin {
	return builtin{ret: ret, args: args, cname: cname}
}

var builtins = map[string]builtin{
	"io.print":    {ret: TInt, args: []Type{TAny}, print: "print"},
	"io.println":  {ret: TInt, args: []Type{TAny}, print: "println"},
	"io.eprint":   {ret: TInt, args: []Type{TAny}, print: "eprint"},
	"io.eprintln": {ret: TInt, args: []Type{TAny}, print: "eprintln"},

	"sys.rune_len": bi(TInt, "lum_utf8_len", TStr),
	"sys.rune_at":  bi(TStr, "lum_utf8_at", TStr, TInt),
	"sys.ord":      bi(TInt, "lum_ord", TStr),
	"sys.chr":      bi(TStr, "lum_chr", TInt),
	"sys.byte_len": bi(TInt, "lum_byte_len", TStr),
	"sys.byte_at":  bi(TInt, "lum_byte_at", TStr, TInt),
	"sys.slice":    bi(TStr, "lum_slice", TStr, TInt, TInt),

	"sys.newline":  bi(TInt, "lum_io_newline"),
	"sys.flush":    bi(TInt, "lum_io_flush"),
	"sys.write":    bi(TInt, "lum_io_write", TStr),
	"sys.readline": bi(TStr, "lum_io_readline"),
	"sys.readint":  bi(TInt, "lum_io_readint"),
	"sys.readchar": bi(TStr, "lum_io_readchar"),
	"sys.readall":  bi(TStr, "lum_io_readall"),

	"sys.getenv":     bi(TStr, "lum_os_getenv", TStr),
	"sys.setenv":     bi(TInt, "lum_os_setenv", TStr, TStr),
	"sys.cwd":        bi(TStr, "lum_os_cwd"),
	"sys.chdir":      bi(TInt, "lum_os_chdir", TStr),
	"sys.pid":        bi(TInt, "lum_os_pid"),
	"sys.ppid":       bi(TInt, "lum_os_ppid"),
	"sys.args_count": bi(TInt, "lum_os_args_count"),
	"sys.arg":        bi(TStr, "lum_os_arg", TInt),
	"sys.exit":       bi(TInt, "lum_os_exit", TInt),
	"sys.sep":        bi(TInt, "lum_os_sep"),
	"sys.is_windows": bi(TBool, "lum_os_is_windows"),
	"sys.hostname":   bi(TStr, "lum_os_hostname"),
	"sys.temp_dir":   bi(TStr, "lum_os_temp_dir"),
	"sys.username":   bi(TStr, "lum_os_username"),
	"sys.arch":       bi(TInt, "lum_os_arch"),

	"sys.now":      bi(TInt, "lum_time_now"),
	"sys.now_ms":   bi(TInt, "lum_time_now_ms"),
	"sys.year":     bi(TInt, "lum_time_year"),
	"sys.month":    bi(TInt, "lum_time_month"),
	"sys.day":      bi(TInt, "lum_time_day"),
	"sys.hour":     bi(TInt, "lum_time_hour"),
	"sys.minute":   bi(TInt, "lum_time_minute"),
	"sys.second":   bi(TInt, "lum_time_second"),
	"sys.weekday":  bi(TInt, "lum_time_weekday"),
	"sys.yearday":  bi(TInt, "lum_time_yearday"),
	"sys.sleep_ms": bi(TInt, "lum_time_sleep_ms", TInt),

	"sys.file_exists": bi(TBool, "lum_file_exists", TStr),
	"sys.file_read":   bi(TStr, "lum_file_read", TStr),
	"sys.file_write":  bi(TInt, "lum_file_write", TStr, TStr),
	"sys.file_append": bi(TInt, "lum_file_append", TStr, TStr),
	"sys.file_remove": bi(TInt, "lum_file_remove", TStr),
	"sys.file_rename": bi(TInt, "lum_file_rename", TStr, TStr),
	"sys.file_size":   bi(TInt, "lum_file_size", TStr),

	"sys.rand_seed": bi(TInt, "lum_rand_seed", TInt),
	"sys.rand_int":  bi(TInt, "lum_rand_int", TInt, TInt),
	"sys.rand_f64":  bi(TFloat, "lum_rand_f64"),

	"sys.system": bi(TInt, "lum_proc_system", TStr),

	"sys.fs_is_dir":  bi(TBool, "lum_fs_is_dir", TStr),
	"sys.fs_is_file": bi(TBool, "lum_fs_is_file", TStr),
	"sys.fs_mkdir":   bi(TInt, "lum_fs_mkdir", TStr),
	"sys.fs_rmdir":   bi(TInt, "lum_fs_rmdir", TStr),
	"sys.fs_list":    bi(TStr, "lum_fs_list", TStr),
	"sys.fs_join":    bi(TStr, "lum_fs_join", TStr, TStr),
	"sys.fs_abs":     bi(TStr, "lum_fs_abs", TStr),

	"sys.msgbox":        bi(TInt, "lum_win_msgbox", TStr, TStr),
	"sys.beep":          bi(TInt, "lum_win_beep", TInt, TInt),
	"sys.tick_ms":       bi(TInt, "lum_win_tick_ms"),
	"sys.console_title": bi(TInt, "lum_win_console_title", TStr),
	"sys.last_error":    bi(TInt, "lum_win_last_error"),
	"sys.win_sleep_ms":  bi(TInt, "lum_win_sleep_ms", TInt),

	"sys.str_cmp": bi(TInt, "lum_str_cmp", TStr, TStr),

	"sys.sin":    bi(TFloat, "lum_math_sin", TFloat),
	"sys.cos":    bi(TFloat, "lum_math_cos", TFloat),
	"sys.tan":    bi(TFloat, "lum_math_tan", TFloat),
	"sys.asin":   bi(TFloat, "lum_math_asin", TFloat),
	"sys.acos":   bi(TFloat, "lum_math_acos", TFloat),
	"sys.atan":   bi(TFloat, "lum_math_atan", TFloat),
	"sys.atan2":  bi(TFloat, "lum_math_atan2", TFloat, TFloat),
	"sys.sqrt_f": bi(TFloat, "lum_math_sqrt", TFloat),
	"sys.pow_f":  bi(TFloat, "lum_math_pow", TFloat, TFloat),
	"sys.fabs":   bi(TFloat, "lum_math_fabs", TFloat),
	"sys.ln":     bi(TFloat, "lum_math_ln", TFloat),
	"sys.log10":  bi(TFloat, "lum_math_log10", TFloat),
	"sys.log2":   bi(TFloat, "lum_math_log2", TFloat),
	"sys.exp":    bi(TFloat, "lum_math_exp", TFloat),
	"sys.floor":  bi(TFloat, "lum_math_floor", TFloat),
	"sys.ceil":   bi(TFloat, "lum_math_ceil", TFloat),
	"sys.round":  bi(TFloat, "lum_math_round", TFloat),
	"sys.trunc":  bi(TFloat, "lum_math_trunc", TFloat),
	"sys.fmod":   bi(TFloat, "lum_math_fmod", TFloat, TFloat),
	"sys.hypot":  bi(TFloat, "lum_math_hypot", TFloat, TFloat),
	"sys.pi":     bi(TFloat, "lum_math_pi"),
	"sys.e":      bi(TFloat, "lum_math_e"),
	"sys.tau":    bi(TFloat, "lum_math_tau"),

	"sys.unix_f":     bi(TFloat, "lum_time_unix_f"),
	"sys.now_str":    bi(TStr, "lum_time_now_str"),
	"sys.utc_offset": bi(TInt, "lum_time_utc_offset"),

	"sys.home_dir":  bi(TStr, "lum_os_home_dir"),
	"sys.is_admin":  bi(TBool, "lum_os_is_admin"),
	"sys.page_size": bi(TInt, "lum_os_page_size"),
	"sys.uname":     bi(TStr, "lum_os_uname"),

	"sys.exec":  bi(TStr, "lum_proc_exec", TStr),
	"sys.which": bi(TStr, "lum_proc_which", TStr),

	"sys.parse_f64":  bi(TFloat, "lum_str_parse_f64", TStr),
	"sys.format_f64": bi(TStr, "lum_str_format_f64", TFloat, TInt),

	"sys.md5":           bi(TStr, "lum_enc_md5", TStr),
	"sys.sha1":          bi(TStr, "lum_enc_sha1", TStr),
	"sys.sha256":        bi(TStr, "lum_enc_sha256", TStr),
	"sys.crc32":         bi(TInt, "lum_enc_crc32", TStr),
	"sys.base64_encode": bi(TStr, "lum_enc_base64_encode", TStr),
	"sys.base64_decode": bi(TStr, "lum_enc_base64_decode", TStr),

	"sys.json_escape":     bi(TStr, "lum_json_escape", TStr),
	"sys.json_unescape":   bi(TStr, "lum_json_unescape", TStr),
	"sys.json_quote":      bi(TStr, "lum_json_quote", TStr),
	"sys.json_parse_f64":  bi(TFloat, "lum_json_parse_f64", TStr),
	"sys.json_parse_int":  bi(TInt, "lum_json_parse_int", TStr),
	"sys.json_parse_bool": bi(TInt, "lum_json_parse_bool", TStr),
	"sys.json_valid":      bi(TBool, "lum_json_valid", TStr),
	"sys.json_type":       bi(TInt, "lum_json_type", TStr),

	"__lum.str_concat": bi(TStr, "lum_str_concat", TStr, TStr),
	"__lum.str_eq":     bi(TBool, "lum_str_eq", TStr, TStr),
	"__lum.utf8_len":   bi(TInt, "lum_utf8_len", TStr),
	"__lum.utf8_at":    bi(TStr, "lum_utf8_at", TStr, TInt),
	"__lum.div":        bi(TInt, "lum_div", TInt, TInt),
	"__lum.mod":        bi(TInt, "lum_mod", TInt, TInt),
	"__lum.input":      bi(TInt, "lum_input"),

	"__lum.print_int":     bi(TInt, "lum_io_print_int", TInt),
	"__lum.print_str":     bi(TInt, "lum_io_print_str", TStr),
	"__lum.print_bool":    bi(TInt, "lum_io_print_bool", TBool),
	"__lum.print_f64":     bi(TInt, "lum_io_print_f64", TFloat),
	"__lum.println_int":   bi(TInt, "lum_io_println_int", TInt),
	"__lum.println_str":   bi(TInt, "lum_io_println_str", TStr),
	"__lum.println_bool":  bi(TInt, "lum_io_println_bool", TBool),
	"__lum.println_f64":   bi(TInt, "lum_io_println_f64", TFloat),
	"__lum.eprint_int":    bi(TInt, "lum_io_eprint_int", TInt),
	"__lum.eprint_str":    bi(TInt, "lum_io_eprint_str", TStr),
	"__lum.eprint_bool":   bi(TInt, "lum_io_eprint_bool", TBool),
	"__lum.eprint_f64":    bi(TInt, "lum_io_eprint_f64", TFloat),
	"__lum.eprintln_int":  bi(TInt, "lum_io_eprintln_int", TInt),
	"__lum.eprintln_str":  bi(TInt, "lum_io_eprintln_str", TStr),
	"__lum.eprintln_bool": bi(TInt, "lum_io_eprintln_bool", TBool),
	"__lum.eprintln_f64":  bi(TInt, "lum_io_eprintln_f64", TFloat),
}

func printSuffix(t Type) string {
	switch t {
	case TInt:
		return "int"
	case TStr:
		return "str"
	case TBool:
		return "bool"
	case TFloat:
		return "f64"
	}
	return ""
}
