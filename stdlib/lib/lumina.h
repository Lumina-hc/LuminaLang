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

#endif
