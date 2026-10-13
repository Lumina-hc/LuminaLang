#ifndef LUMINA_H
#define LUMINA_H

#define _CRT_SECURE_NO_WARNINGS

#include <stdint.h>
#include <stddef.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

char *lum_dup(const char *s);
void *lum_alloc(size_t n);
void lum_arena_free_all(void);
int64_t lum_utf8_len(const char *s);
const char *lum_utf8_at(const char *s, int64_t idx);
int64_t lum_ord(const char *s);
const char *lum_chr(int64_t cp);
int64_t lum_byte_len(const char *s);
int64_t lum_byte_at(const char *s, int64_t i);
const char *lum_slice(const char *s, int64_t start, int64_t end);

int64_t lum_slice_len(const void *s);
void *lum_slice_new(int64_t cap);
int64_t lum_slice_get_int(const void *s, int64_t i);
double lum_slice_get_f64(const void *s, int64_t i);
int64_t lum_slice_get_bool(const void *s, int64_t i);
const char *lum_slice_get_str(const void *s, int64_t i);
void lum_slice_set_int(void *s, int64_t i, int64_t v);
void lum_slice_set_f64(void *s, int64_t i, double v);
void lum_slice_set_bool(void *s, int64_t i, int64_t v);
void lum_slice_set_str(void *s, int64_t i, const char *v);
void *lum_slice_push_int(void *s, int64_t v);
void *lum_slice_push_f64(void *s, double v);
void *lum_slice_push_bool(void *s, int64_t v);
void *lum_slice_push_str(void *s, const char *v);

int lum_str_cmp(const char *a, const char *b);

/* math (f64) */
double lum_math_sin(double x);
double lum_math_cos(double x);
double lum_math_tan(double x);
double lum_math_asin(double x);
double lum_math_acos(double x);
double lum_math_atan(double x);
double lum_math_atan2(double y, double x);
double lum_math_sqrt(double x);
double lum_math_pow(double x, double y);
double lum_math_fabs(double x);
double lum_math_ln(double x);
double lum_math_log10(double x);
double lum_math_log2(double x);
double lum_math_exp(double x);
double lum_math_floor(double x);
double lum_math_ceil(double x);
double lum_math_round(double x);
double lum_math_trunc(double x);
double lum_math_fmod(double x, double y);
double lum_math_hypot(double x, double y);
double lum_math_pi(void);
double lum_math_e(void);
double lum_math_tau(void);

/* time */
double lum_time_unix_f(void);
const char *lum_time_now_str(void);
int64_t lum_time_utc_offset(void);

/* os */
const char *lum_os_home_dir(void);
int lum_os_is_admin(void);
int64_t lum_os_page_size(void);
const char *lum_os_uname(void);

/* process */
const char *lum_proc_exec(const char *cmd);
const char *lum_proc_which(const char *prog);

/* strconv */
double lum_str_parse_f64(const char *s);
const char *lum_str_format_f64(double v, int64_t prec);

/* encoding */
const char *lum_enc_md5(const char *s);
const char *lum_enc_sha1(const char *s);
const char *lum_enc_sha256(const char *s);
int64_t lum_enc_crc32(const char *s);
const char *lum_enc_base64_encode(const char *s);
const char *lum_enc_base64_decode(const char *s);

/* json */
const char *lum_json_escape(const char *s);
const char *lum_json_unescape(const char *s);
const char *lum_json_quote(const char *s);
double lum_json_parse_f64(const char *s);
int64_t lum_json_parse_int(const char *s);
int64_t lum_json_parse_bool(const char *s);
int lum_json_valid(const char *s);
int64_t lum_json_type(const char *s);

#endif
