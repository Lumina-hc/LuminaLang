#include "lumina.h"

#ifdef _WIN32
#define LUM_POPEN _popen
#define LUM_PCLOSE _pclose
#else
#include <unistd.h>
#define LUM_POPEN popen
#define LUM_PCLOSE pclose
#endif

int64_t lum_proc_system(const char *cmd) {
    if (!cmd) return -1;
    return (int64_t)system(cmd);
}

const char *lum_proc_exec(const char *cmd) {
    if (!cmd) return lum_dup("");
    FILE *p = LUM_POPEN(cmd, "r");
    if (!p) return lum_dup("");
    size_t cap = 4096, len = 0;
    char *buf = (char *)malloc(cap);
    if (!buf) { LUM_PCLOSE(p); return lum_dup(""); }
    size_t n;
    while ((n = fread(buf + len, 1, cap - len - 1, p)) > 0) {
        len += n;
        if (len + 1 >= cap) {
            cap *= 2;
            char *nb = (char *)realloc(buf, cap);
            if (!nb) break;
            buf = nb;
        }
    }
    buf[len] = 0;
    LUM_PCLOSE(p);
    const char *r = lum_dup(buf);
    free(buf);
    return r;
}

const char *lum_proc_which(const char *prog) {
    if (!prog || !*prog) return lum_dup("");
#ifdef _WIN32
    const char *path = getenv("PATH");
    if (!path) return lum_dup("");
    char *copy = (char *)malloc(strlen(path) + 1);
    if (!copy) return lum_dup("");
    strcpy(copy, path);
    char *tok = strtok(copy, ";");
    while (tok) {
        char buf[4096];
        snprintf(buf, sizeof(buf), "%s\\%s.exe", tok, prog);
        FILE *f = fopen(buf, "r");
        if (f) { fclose(f); free(copy); return lum_dup(buf); }
        snprintf(buf, sizeof(buf), "%s\\%s", tok, prog);
        f = fopen(buf, "r");
        if (f) { fclose(f); free(copy); return lum_dup(buf); }
        tok = strtok(NULL, ";");
    }
    free(copy);
    return lum_dup("");
#else
    const char *path = getenv("PATH");
    if (!path) return lum_dup("");
    char *copy = (char *)malloc(strlen(path) + 1);
    if (!copy) return lum_dup("");
    strcpy(copy, path);
    char *tok = strtok(copy, ":");
    while (tok) {
        char buf[4096];
        snprintf(buf, sizeof(buf), "%s/%s", tok, prog);
        if (access(buf, X_OK) == 0) { free(copy); return lum_dup(buf); }
        tok = strtok(NULL, ":");
    }
    free(copy);
    return lum_dup("");
#endif
}
