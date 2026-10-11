#include "lumina.h"

#ifdef _WIN32
#include <windows.h>
#include <direct.h>
#define LUM_MKDIR(p) _mkdir(p)
#define LUM_RMDIR(p) _rmdir(p)
#else
#include <sys/stat.h>
#include <sys/types.h>
#include <unistd.h>
#include <dirent.h>
#define LUM_MKDIR(p) mkdir(p, 0777)
#define LUM_RMDIR(p) rmdir(p)
#endif

int64_t lum_fs_is_dir(const char *path) {
    if (!path) return 0;
#ifdef _WIN32
    DWORD a = GetFileAttributesA(path);
    return (a != INVALID_FILE_ATTRIBUTES && (a & FILE_ATTRIBUTE_DIRECTORY)) ? 1 : 0;
#else
    struct stat st;
    if (stat(path, &st) != 0) return 0;
    return S_ISDIR(st.st_mode) ? 1 : 0;
#endif
}

int64_t lum_fs_is_file(const char *path) {
    if (!path) return 0;
#ifdef _WIN32
    DWORD a = GetFileAttributesA(path);
    return (a != INVALID_FILE_ATTRIBUTES && !(a & FILE_ATTRIBUTE_DIRECTORY)) ? 1 : 0;
#else
    struct stat st;
    if (stat(path, &st) != 0) return 0;
    return S_ISREG(st.st_mode) ? 1 : 0;
#endif
}

int64_t lum_fs_mkdir(const char *path) {
    if (!path) return -1;
    return LUM_MKDIR(path) == 0 ? 0 : -1;
}

int64_t lum_fs_rmdir(const char *path) {
    if (!path) return -1;
    return LUM_RMDIR(path) == 0 ? 0 : -1;
}

const char *lum_fs_join(const char *a, const char *b) {
    if (!a) a = "";
    if (!b) b = "";
    size_t na = strlen(a), nb = strlen(b);
#ifdef _WIN32
    char sep = '\\';
    int have = na > 0 && (a[na - 1] == '\\' || a[na - 1] == '/');
#else
    char sep = '/';
    int have = na > 0 && a[na - 1] == '/';
#endif
    char *r = (char *)lum_alloc(na + nb + 2);
    memcpy(r, a, na);
    size_t k = na;
    if (na > 0 && !have) r[k++] = sep;
    memcpy(r + k, b, nb);
    r[k + nb] = 0;
    return r;
}

const char *lum_fs_abs(const char *path) {
    if (!path) return lum_dup("");
#ifdef _WIN32
    char buf[4096];
    if (!_fullpath(buf, path, sizeof(buf))) return lum_dup("");
    return lum_dup(buf);
#else
    char *r = realpath(path, NULL);
    if (!r) return lum_dup("");
    const char *s = lum_dup(r);
    free(r);
    return s;
#endif
}

static void lum_fs_append(char **buf, size_t *len, size_t *cap, const char *name) {
    size_t n = strlen(name);
    if (*len + n + 2 > *cap) {
        if (*cap == 0) *cap = 256;
        while (*len + n + 2 > *cap) *cap *= 2;
        *buf = (char *)realloc(*buf, *cap);
    }
    if (*len > 0) (*buf)[(*len)++] = '\n';
    memcpy(*buf + *len, name, n);
    *len += n;
    (*buf)[*len] = 0;
}

const char *lum_fs_list(const char *path) {
    char *buf = NULL;
    size_t len = 0, cap = 0;
#ifdef _WIN32
    if (!path) return lum_dup("");
    char pat[4096];
    snprintf(pat, sizeof(pat), "%s\\*", path);
    WIN32_FIND_DATAA fd;
    HANDLE h = FindFirstFileA(pat, &fd);
    if (h == INVALID_HANDLE_VALUE) return lum_dup("");
    do {
        if (strcmp(fd.cFileName, ".") == 0 || strcmp(fd.cFileName, "..") == 0) continue;
        lum_fs_append(&buf, &len, &cap, fd.cFileName);
    } while (FindNextFileA(h, &fd));
    FindClose(h);
#else
    DIR *d = (path && *path) ? opendir(path) : opendir(".");
    if (!d) return lum_dup("");
    struct dirent *e;
    while ((e = readdir(d))) {
        if (strcmp(e->d_name, ".") == 0 || strcmp(e->d_name, "..") == 0) continue;
        lum_fs_append(&buf, &len, &cap, e->d_name);
    }
    closedir(d);
#endif
    if (!buf) return lum_dup("");
    return buf;
}
