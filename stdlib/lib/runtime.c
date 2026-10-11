#include "lumina.h"

int64_t lumf_main(void);

int64_t lum_argc_g = 0;
char **lum_argv_g = NULL;

typedef struct LumBlock {
    struct LumBlock *next;
    size_t used;
    size_t cap;
    char *data;
} LumBlock;

static LumBlock *lum_arena = NULL;

void *lum_alloc(size_t n) {
    n = (n + 15u) & ~(size_t)15u;
    if (!lum_arena || lum_arena->used + n > lum_arena->cap) {
        size_t cap = n > 65536u ? n : 65536u;
        LumBlock *b = (LumBlock *)malloc(sizeof(LumBlock) + cap);
        if (!b) {
            fputs("runtime error: out of memory\n", stderr);
            exit(1);
        }
        b->data = (char *)(b + 1);
        b->cap = cap;
        b->used = 0;
        b->next = lum_arena;
        lum_arena = b;
    }
    void *p = lum_arena->data + lum_arena->used;
    lum_arena->used += n;
    return p;
}

void lum_arena_free_all(void) {
    LumBlock *b = lum_arena;
    while (b) {
        LumBlock *n = b->next;
        free(b);
        b = n;
    }
    lum_arena = NULL;
}

int main(int argc, char **argv) {
    lum_argc_g = (int64_t)argc;
    lum_argv_g = argv;
    int64_t rc = lumf_main();
    fflush(stdout);
    lum_arena_free_all();
    return (int)rc;
}

char *lum_dup(const char *s) {
    if (!s) s = "";
    size_t n = strlen(s);
    char *r = (char *)lum_alloc(n + 1);
    memcpy(r, s, n + 1);
    return r;
}

int lum_str_eq(const char *a, const char *b) {
    if (!a) a = "";
    if (!b) b = "";
    return strcmp(a, b) == 0;
}

const char *lum_str_concat(const char *a, const char *b) {
    if (!a) a = "";
    if (!b) b = "";
    size_t la = strlen(a), lb = strlen(b);
    char *r = (char *)lum_alloc(la + lb + 1);
    memcpy(r, a, la);
    memcpy(r + la, b, lb + 1);
    return r;
}

static int utf8_seq_len(unsigned char c) {
    if (c < 0x80) return 1;
    if ((c >> 5) == 0x6) return 2;
    if ((c >> 4) == 0xE) return 3;
    if ((c >> 3) == 0x1E) return 4;
    return 1;
}

int64_t lum_utf8_len(const char *s) {
    if (!s) return 0;
    const unsigned char *p = (const unsigned char *)s;
    int64_t n = 0;
    while (*p) { p += utf8_seq_len(*p); n++; }
    return n;
}

const char *lum_utf8_at(const char *s, int64_t idx) {
    if (!s) return lum_dup("");
    const unsigned char *p = (const unsigned char *)s;
    while (idx > 0 && *p) { p += utf8_seq_len(*p); idx--; }
    if (!*p) return lum_dup("");
    int len = utf8_seq_len(*p);
    char *r = (char *)lum_alloc((size_t)len + 1);
    memcpy(r, p, (size_t)len);
    r[len] = 0;
    return r;
}

int64_t lum_str_len_bytes(const char *s) {
    return s ? (int64_t)strlen(s) : 0;
}

int64_t lum_ord(const char *s) {
    if (!s || !*s) return 0;
    const unsigned char *p = (const unsigned char *)s;
    int len = utf8_seq_len(*p);
    int64_t cp = 0;
    switch (len) {
        case 1: cp = p[0]; break;
        case 2: cp = (int64_t)(p[0] & 0x1F) << 6 | (p[1] & 0x3F); break;
        case 3: cp = (int64_t)(p[0] & 0x0F) << 12 | (int64_t)(p[1] & 0x3F) << 6 | (p[2] & 0x3F); break;
        case 4: cp = (int64_t)(p[0] & 0x07) << 18 | (int64_t)(p[1] & 0x3F) << 12 |
                    (int64_t)(p[2] & 0x3F) << 6 | (p[3] & 0x3F); break;
        default: cp = p[0]; break;
    }
    return cp;
}

const char *lum_chr(int64_t cp) {
    unsigned char b[5];
    int n = 0;
    if (cp < 0) cp = 0xFFFD;
    if (cp < 0x80) {
        b[n++] = (unsigned char)cp;
    } else if (cp < 0x800) {
        b[n++] = (unsigned char)(0xC0 | (cp >> 6));
        b[n++] = (unsigned char)(0x80 | (cp & 0x3F));
    } else if (cp < 0x10000) {
        b[n++] = (unsigned char)(0xE0 | (cp >> 12));
        b[n++] = (unsigned char)(0x80 | ((cp >> 6) & 0x3F));
        b[n++] = (unsigned char)(0x80 | (cp & 0x3F));
    } else if (cp < 0x110000) {
        b[n++] = (unsigned char)(0xF0 | (cp >> 18));
        b[n++] = (unsigned char)(0x80 | ((cp >> 12) & 0x3F));
        b[n++] = (unsigned char)(0x80 | ((cp >> 6) & 0x3F));
        b[n++] = (unsigned char)(0x80 | (cp & 0x3F));
    } else {
        cp = 0xFFFD;
        b[n++] = (unsigned char)(0xE0 | (cp >> 12));
        b[n++] = (unsigned char)(0x80 | ((cp >> 6) & 0x3F));
        b[n++] = (unsigned char)(0x80 | (cp & 0x3F));
    }
    b[n] = 0;
    char *r = (char *)lum_alloc((size_t)n + 1);
    memcpy(r, b, (size_t)n + 1);
    return r;
}

int64_t lum_byte_len(const char *s) { return s ? (int64_t)strlen(s) : 0; }

int64_t lum_byte_at(const char *s, int64_t i) {
    if (!s || i < 0) return -1;
    size_t n = strlen(s);
    if ((size_t)i >= n) return -1;
    return (int64_t)(unsigned char)s[i];
}

const char *lum_slice(const char *s, int64_t start, int64_t end) {
    if (!s) s = "";
    int64_t n = (int64_t)strlen(s);
    if (start < 0) start = 0;
    if (start > n) start = n;
    if (end < start) end = start;
    if (end > n) end = n;
    int64_t len = end - start;
    char *r = (char *)lum_alloc((size_t)len + 1);
    memcpy(r, s + start, (size_t)len);
    r[len] = 0;
    return r;
}

int lum_str_cmp(const char *a, const char *b) {
    if (!a) a = "";
    if (!b) b = "";
    return strcmp(a, b);
}

int64_t lum_div(int64_t a, int64_t b) {
    if (b == 0) {
        fputs("runtime error: division by zero\n", stderr);
        exit(1);
    }
    int64_t q = a / b;
    if ((a % b != 0) && ((a < 0) != (b < 0))) q--;
    return q;
}

int64_t lum_mod(int64_t a, int64_t b) {
    if (b == 0) {
        fputs("runtime error: modulo by zero\n", stderr);
        exit(1);
    }
    int64_t r = a % b;
    if (r != 0 && ((r < 0) != (b < 0))) r += b;
    return r;
}

int lum_truthy(int64_t v) { return v != 0; }

static int64_t lum_promote(const char *s) {
    if (!s) return 0;
    char *end = NULL;
    long long v = strtoll(s, &end, 10);
    if (end == s) return 0;
    return (int64_t)v;
}

int64_t lum_input(void) {
    char buf[64];
    if (scanf("%63s", buf) != 1) return 0;
    return lum_promote(buf);
}

typedef struct {
    int64_t len;
    int64_t cap;
    int64_t data[1];
} LumSlice;

static void lum_slice_oob(int64_t i, int64_t len) {
    fprintf(stderr, "runtime error: slice index %lld out of range (length %lld)\n",
            (long long)i, (long long)len);
    exit(1);
}

int64_t lum_slice_len(const void *s) {
    if (!s) return 0;
    return ((const LumSlice *)s)->len;
}

void *lum_slice_new(int64_t len) {
    if (len < 0) len = 0;
    int64_t cap = len < 1 ? 1 : len;
    LumSlice *s = (LumSlice *)lum_alloc(offsetof(LumSlice, data) +
                                        (size_t)cap * sizeof(int64_t));
    s->len = len;
    s->cap = cap;
    memset(s->data, 0, (size_t)cap * sizeof(int64_t));
    return s;
}

static LumSlice *lum_slice_grow(void *sp) {
    LumSlice *s = (LumSlice *)sp;
    if (!s) {
        s = (LumSlice *)lum_slice_new(1);
        return s;
    }
    if (s->len >= s->cap) {
        int64_t ncap = s->cap * 2;
        LumSlice *ns = (LumSlice *)lum_alloc(offsetof(LumSlice, data) +
                                             (size_t)ncap * sizeof(int64_t));
        ns->len = s->len;
        ns->cap = ncap;
        memcpy(ns->data, s->data, (size_t)s->len * sizeof(int64_t));
        return ns;
    }
    return s;
}

static int64_t lum_slice_get_raw(const void *sp, int64_t i) {
    const LumSlice *s = (const LumSlice *)sp;
    if (!s || i < 0 || i >= s->len) lum_slice_oob(i, s ? s->len : 0);
    return s->data[i];
}

int64_t lum_slice_get_int(const void *s, int64_t i) {
    return lum_slice_get_raw(s, i);
}

double lum_slice_get_f64(const void *s, int64_t i) {
    int64_t bits = lum_slice_get_raw(s, i);
    double d;
    memcpy(&d, &bits, sizeof(double));
    return d;
}

int64_t lum_slice_get_bool(const void *s, int64_t i) {
    return lum_slice_get_raw(s, i) != 0;
}

const char *lum_slice_get_str(const void *s, int64_t i) {
    const char *p = (const char *)(intptr_t)lum_slice_get_raw(s, i);
    return p ? p : "";
}

static LumSlice *lum_slice_set_raw(void *sp, int64_t i, int64_t v) {
    LumSlice *s = (LumSlice *)sp;
    if (!s || i < 0 || i >= s->len) lum_slice_oob(i, s ? s->len : 0);
    s->data[i] = v;
    return s;
}

void lum_slice_set_int(void *s, int64_t i, int64_t v) {
    (void)lum_slice_set_raw(s, i, v);
}

void lum_slice_set_f64(void *s, int64_t i, double v) {
    int64_t bits;
    memcpy(&bits, &v, sizeof(double));
    (void)lum_slice_set_raw(s, i, bits);
}

void lum_slice_set_bool(void *s, int64_t i, int64_t v) {
    (void)lum_slice_set_raw(s, i, v != 0);
}

void lum_slice_set_str(void *s, int64_t i, const char *v) {
    (void)lum_slice_set_raw(s, i, (int64_t)(intptr_t)(v ? v : ""));
}

void *lum_slice_push_int(void *s, int64_t v) {
    LumSlice *ns = lum_slice_grow(s);
    ns->data[ns->len++] = v;
    return ns;
}

void *lum_slice_push_f64(void *s, double v) {
    LumSlice *ns = lum_slice_grow(s);
    int64_t bits;
    memcpy(&bits, &v, sizeof(double));
    ns->data[ns->len++] = bits;
    return ns;
}

void *lum_slice_push_bool(void *s, int64_t v) {
    LumSlice *ns = lum_slice_grow(s);
    ns->data[ns->len++] = v != 0;
    return ns;
}

void *lum_slice_push_str(void *s, const char *v) {
    LumSlice *ns = lum_slice_grow(s);
    ns->data[ns->len++] = (int64_t)(intptr_t)(v ? v : "");
    return ns;
}
