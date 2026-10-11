#include "lumina.h"

int lum_file_exists(const char *path) {
    if (!path) return 0;
    FILE *f = fopen(path, "rb");
    if (f) { fclose(f); return 1; }
    return 0;
}

const char *lum_file_read(const char *path) {
    if (!path) return lum_dup("");
    FILE *f = fopen(path, "rb");
    if (!f) return lum_dup("");
    fseek(f, 0, SEEK_END);
    long n = ftell(f);
    fseek(f, 0, SEEK_SET);
    if (n < 0) { fclose(f); return lum_dup(""); }
    char *r = (char *)lum_alloc((size_t)n + 1);
    size_t got = fread(r, 1, (size_t)n, f);
    fclose(f);
    r[got] = 0;
    return r;
}

int64_t lum_file_write(const char *path, const char *data) {
    if (!path) return -1;
    FILE *f = fopen(path, "wb");
    if (!f) return -1;
    if (data) fwrite(data, 1, strlen(data), f);
    fclose(f);
    return 0;
}

int64_t lum_file_append(const char *path, const char *data) {
    if (!path) return -1;
    FILE *f = fopen(path, "ab");
    if (!f) return -1;
    if (data) fwrite(data, 1, strlen(data), f);
    fclose(f);
    return 0;
}

int64_t lum_file_remove(const char *path) {
    if (!path) return -1;
    return remove(path) == 0 ? 0 : -1;
}

int64_t lum_file_rename(const char *from, const char *to) {
    if (!from || !to) return -1;
    return rename(from, to) == 0 ? 0 : -1;
}

int64_t lum_file_size(const char *path) {
    if (!path) return -1;
    FILE *f = fopen(path, "rb");
    if (!f) return -1;
    fseek(f, 0, SEEK_END);
    long n = ftell(f);
    fclose(f);
    return n;
}
