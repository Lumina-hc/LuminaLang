#include "lumina.h"

static const char *lum_jskip(const char *p) {
    while (*p == ' ' || *p == '\t' || *p == '\n' || *p == '\r') p++;
    return p;
}

const char *lum_json_escape(const char *s) {
    if (!s) s = "";
    size_t n = strlen(s);
    char *r = (char *)lum_alloc(n * 6 + 1);
    size_t o = 0;
    static const char *hex = "0123456789abcdef";
    for (size_t i = 0; i < n; i++) {
        unsigned char c = (unsigned char)s[i];
        switch (c) {
            case '"':  r[o++] = '\\'; r[o++] = '"'; break;
            case '\\': r[o++] = '\\'; r[o++] = '\\'; break;
            case '\n': r[o++] = '\\'; r[o++] = 'n'; break;
            case '\r': r[o++] = '\\'; r[o++] = 'r'; break;
            case '\t': r[o++] = '\\'; r[o++] = 't'; break;
            case '\b': r[o++] = '\\'; r[o++] = 'b'; break;
            case '\f': r[o++] = '\\'; r[o++] = 'f'; break;
            default:
                if (c < 0x20) {
                    r[o++] = '\\'; r[o++] = 'u'; r[o++] = '0'; r[o++] = '0';
                    r[o++] = hex[(c >> 4) & 15];
                    r[o++] = hex[c & 15];
                } else {
                    r[o++] = (char)c;
                }
        }
    }
    r[o] = 0;
    return r;
}

const char *lum_json_quote(const char *s) {
    const char *body = lum_json_escape(s);
    size_t n = strlen(body);
    char *r = (char *)lum_alloc(n + 3);
    r[0] = '"';
    memcpy(r + 1, body, n);
    r[n + 1] = '"';
    r[n + 2] = 0;
    return r;
}

static int lum_hexval(int c) {
    if (c >= '0' && c <= '9') return c - '0';
    if (c >= 'a' && c <= 'f') return c - 'a' + 10;
    if (c >= 'A' && c <= 'F') return c - 'A' + 10;
    return -1;
}

static size_t lum_utf8_put(char *out, uint32_t cp) {
    if (cp < 0x80) {
        out[0] = (char)cp;
        return 1;
    } else if (cp < 0x800) {
        out[0] = (char)(0xC0 | (cp >> 6));
        out[1] = (char)(0x80 | (cp & 0x3F));
        return 2;
    } else if (cp < 0x10000) {
        out[0] = (char)(0xE0 | (cp >> 12));
        out[1] = (char)(0x80 | ((cp >> 6) & 0x3F));
        out[2] = (char)(0x80 | (cp & 0x3F));
        return 3;
    } else {
        out[0] = (char)(0xF0 | (cp >> 18));
        out[1] = (char)(0x80 | ((cp >> 12) & 0x3F));
        out[2] = (char)(0x80 | ((cp >> 6) & 0x3F));
        out[3] = (char)(0x80 | (cp & 0x3F));
        return 4;
    }
}

const char *lum_json_unescape(const char *s) {
    if (!s) s = "";
    size_t n = strlen(s);
    char *r = (char *)lum_alloc(n + 1);
    size_t o = 0;
    for (size_t i = 0; i < n; i++) {
        char c = s[i];
        if (c != '\\') { r[o++] = c; continue; }
        i++;
        if (i >= n) break;
        char e = s[i];
        switch (e) {
            case 'n': r[o++] = '\n'; break;
            case 'r': r[o++] = '\r'; break;
            case 't': r[o++] = '\t'; break;
            case 'b': r[o++] = '\b'; break;
            case 'f': r[o++] = '\f'; break;
            case '/': r[o++] = '/'; break;
            case '"': r[o++] = '"'; break;
            case '\\': r[o++] = '\\'; break;
            case 'u': {
                if (i + 4 >= n) { break; }
                uint32_t cp = 0;
                int ok = 1;
                for (int k = 1; k <= 4; k++) {
                    int hv = lum_hexval(s[i + k]);
                    if (hv < 0) { ok = 0; break; }
                    cp = (cp << 4) | (uint32_t)hv;
                }
                if (!ok) break;
                i += 4;
                if (cp >= 0xD800 && cp <= 0xDBFF && i + 6 < n &&
                    s[i + 1] == '\\' && s[i + 2] == 'u') {
                    uint32_t lo = 0;
                    int ok2 = 1;
                    for (int k = 3; k <= 6; k++) {
                        int hv = lum_hexval(s[i + k]);
                        if (hv < 0) { ok2 = 0; break; }
                        lo = (lo << 4) | (uint32_t)hv;
                    }
                    if (ok2 && lo >= 0xDC00 && lo <= 0xDFFF) {
                        cp = 0x10000 + ((cp - 0xD800) << 10) + (lo - 0xDC00);
                        i += 6;
                    }
                }
                o += lum_utf8_put(r + o, cp);
                break;
            }
            default: r[o++] = e; break;
        }
    }
    r[o] = 0;
    return r;
}

double lum_json_parse_f64(const char *s) {
    if (!s) return 0.0;
    return strtod(lum_jskip(s), NULL);
}

int64_t lum_json_parse_int(const char *s) {
    if (!s) return 0;
    return (int64_t)strtoll(lum_jskip(s), NULL, 10);
}

int64_t lum_json_parse_bool(const char *s) {
    if (!s) return -1;
    const char *p = lum_jskip(s);
    if (strncmp(p, "true", 4) == 0) return 1;
    if (strncmp(p, "false", 5) == 0) return 0;
    return -1;
}

static const char *lum_json_str_end(const char *p) {
    p++;
    while (*p) {
        if (*p == '\\') { p++; if (*p) p++; }
        else if (*p == '"') return p + 1;
        else p++;
    }
    return NULL;
}

static const char *lum_json_number_end(const char *p) {
    if (*p == '-') p++;
    if (*p == '0') { p++; }
    else if (*p >= '1' && *p <= '9') { while (*p >= '0' && *p <= '9') p++; }
    else return NULL;
    if (*p == '.') {
        p++;
        if (!(*p >= '0' && *p <= '9')) return NULL;
        while (*p >= '0' && *p <= '9') p++;
    }
    if (*p == 'e' || *p == 'E') {
        p++;
        if (*p == '+' || *p == '-') p++;
        if (!(*p >= '0' && *p <= '9')) return NULL;
        while (*p >= '0' && *p <= '9') p++;
    }
    return p;
}

static const char *lum_json_value_end(const char *p, int depth) {
    if (depth > 128) return NULL;
    p = lum_jskip(p);
    if (*p == '"') return lum_json_str_end(p);
    if (*p == '{') {
        p = lum_jskip(p + 1);
        if (*p == '}') return p + 1;
        for (;;) {
            p = lum_jskip(p);
            if (*p != '"') return NULL;
            p = lum_json_str_end(p);
            if (!p) return NULL;
            p = lum_jskip(p);
            if (*p != ':') return NULL;
            p = lum_json_value_end(p + 1, depth + 1);
            if (!p) return NULL;
            p = lum_jskip(p);
            if (*p == ',') { p++; continue; }
            if (*p == '}') return p + 1;
            return NULL;
        }
    }
    if (*p == '[') {
        p = lum_jskip(p + 1);
        if (*p == ']') return p + 1;
        for (;;) {
            p = lum_json_value_end(p, depth + 1);
            if (!p) return NULL;
            p = lum_jskip(p);
            if (*p == ',') { p++; continue; }
            if (*p == ']') return p + 1;
            return NULL;
        }
    }
    if (strncmp(p, "true", 4) == 0) return p + 4;
    if (strncmp(p, "false", 5) == 0) return p + 5;
    if (strncmp(p, "null", 4) == 0) return p + 4;
    return lum_json_number_end(p);
}

int lum_json_valid(const char *s) {
    if (!s) return 0;
    const char *p = lum_json_value_end(s, 0);
    if (!p) return 0;
    p = lum_jskip(p);
    return *p == 0;
}

int64_t lum_json_type(const char *s) {
    if (!s) return 0;
    const char *p = lum_jskip(s);
    switch (*p) {
        case '"': return 2;
        case '{': return 4;
        case '[': return 3;
        case 't': return strncmp(p, "true", 4) == 0 ? 5 : 0;
        case 'f': return strncmp(p, "false", 5) == 0 ? 6 : 0;
        case 'n': return strncmp(p, "null", 4) == 0 ? 7 : 0;
        default:  return (*p == '-' || (*p >= '0' && *p <= '9')) ? 1 : 0;
    }
}
