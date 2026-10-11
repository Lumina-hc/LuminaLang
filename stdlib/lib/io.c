#include "lumina.h"

int64_t lum_io_print_int(int64_t v)   { printf("%lld", (long long)v); return 0; }
int64_t lum_io_println_int(int64_t v) { printf("%lld\n", (long long)v); return 0; }
int64_t lum_io_print_bool(int v)      { fputs(v ? "true" : "false", stdout); return 0; }
int64_t lum_io_println_bool(int v)    { fputs(v ? "true" : "false", stdout); fputc('\n', stdout); return 0; }
int64_t lum_io_print_str(const char *s)   { fputs(s ? s : "", stdout); return 0; }
int64_t lum_io_println_str(const char *s) { fputs(s ? s : "", stdout); fputc('\n', stdout); return 0; }
int64_t lum_io_print_f64(double v)   { printf("%g", v); return 0; }
int64_t lum_io_println_f64(double v) { printf("%g\n", v); return 0; }

int64_t lum_io_eprint_int(int64_t v)   { fprintf(stderr, "%lld", (long long)v); return 0; }
int64_t lum_io_eprintln_int(int64_t v) { fprintf(stderr, "%lld\n", (long long)v); return 0; }
int64_t lum_io_eprint_bool(int v)      { fputs(v ? "true" : "false", stderr); return 0; }
int64_t lum_io_eprintln_bool(int v)    { fputs(v ? "true" : "false", stderr); fputc('\n', stderr); return 0; }
int64_t lum_io_eprint_str(const char *s)   { fputs(s ? s : "", stderr); return 0; }
int64_t lum_io_eprintln_str(const char *s) { fputs(s ? s : "", stderr); fputc('\n', stderr); return 0; }
int64_t lum_io_eprint_f64(double v)   { fprintf(stderr, "%g", v); return 0; }
int64_t lum_io_eprintln_f64(double v) { fprintf(stderr, "%g\n", v); return 0; }

int64_t lum_io_newline(void) { fputc('\n', stdout); return 0; }

int64_t lum_io_flush(void) { fflush(stdout); fflush(stderr); return 0; }

int64_t lum_io_write(const char *s) { fputs(s ? s : "", stdout); return 0; }

const char *lum_io_readline(void) {
    static char *buf = NULL;
    static size_t cap = 0;
    size_t n = 0;
    int c;
    if (!buf) { cap = 128; buf = (char *)malloc(cap); }
    while ((c = fgetc(stdin)) != EOF && c != '\n') {
        if (n + 1 >= cap) { cap *= 2; buf = (char *)realloc(buf, cap); }
        buf[n++] = (char)c;
    }
    if (c == EOF && n == 0) return lum_dup("");
    if (n > 0 && buf[n - 1] == '\r') n--;
    buf[n] = 0;
    return lum_dup(buf);
}

int64_t lum_io_readint(void) {
    char buf[64];
    if (scanf("%63s", buf) != 1) return 0;
    return (int64_t)strtoll(buf, NULL, 10);
}

const char *lum_io_readchar(void) {
    int c = fgetc(stdin);
    if (c == EOF) return lum_dup("");
    char b[2];
    b[0] = (char)c;
    b[1] = 0;
    return lum_dup(b);
}

const char *lum_io_readall(void) {
    size_t cap = 4096, n = 0;
    char *r = (char *)malloc(cap);
    size_t got;
    while ((got = fread(r + n, 1, cap - n, stdin)) > 0) {
        n += got;
        if (n == cap) { cap *= 2; r = (char *)realloc(r, cap); }
    }
    r[n] = 0;
    return r;
}
