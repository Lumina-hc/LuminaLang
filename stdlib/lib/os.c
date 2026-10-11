#include "lumina.h"

#ifdef _WIN32
#include <windows.h>
#include <direct.h>
#include <process.h>
#define LUM_GETCWD _getcwd
#define LUM_GETPID _getpid
#else
#include <unistd.h>
#include <limits.h>
#define LUM_GETCWD getcwd
#define LUM_GETPID getpid
#endif

extern int64_t lum_argc_g;
extern char **lum_argv_g;

const char *lum_os_getenv(const char *key) {
    if (!key) return lum_dup("");
    const char *v = getenv(key);
    return lum_dup(v ? v : "");
}

int64_t lum_os_setenv(const char *key, const char *val) {
    if (!key) return -1;
#ifdef _WIN32
    char buf[4096];
    snprintf(buf, sizeof(buf), "%s=%s", key, val ? val : "");
    return _putenv(buf) == 0 ? 0 : -1;
#else
    return setenv(key, val ? val : "", 1) == 0 ? 0 : -1;
#endif
}

const char *lum_os_cwd(void) {
    char *d = LUM_GETCWD(NULL, 0);
    if (!d) return lum_dup("");
    const char *r = lum_dup(d);
    free(d);
    return r;
}

int64_t lum_os_chdir(const char *path) {
    if (!path) return -1;
#ifdef _WIN32
    return _chdir(path) == 0 ? 0 : -1;
#else
    return chdir(path) == 0 ? 0 : -1;
#endif
}

int64_t lum_os_pid(void) { return (int64_t)LUM_GETPID(); }

int64_t lum_os_ppid(void) {
#ifdef _WIN32
    return 0;
#else
    return (int64_t)getppid();
#endif
}

int64_t lum_os_args_count(void) { return lum_argc_g; }

const char *lum_os_arg(int64_t i) {
    if (i < 0 || i >= lum_argc_g || !lum_argv_g) return lum_dup("");
    return lum_dup(lum_argv_g[i]);
}

int64_t lum_os_exit(int64_t code) {
    fflush(stdout);
    exit((int)code);
    return 0;
}

int64_t lum_os_sep(void) {
#ifdef _WIN32
    return (int64_t)'\\';
#else
    return (int64_t)'/';
#endif
}

int lum_os_is_windows(void) {
#ifdef _WIN32
    return 1;
#else
    return 0;
#endif
}

const char *lum_os_hostname(void) {
    char buf[256];
#ifdef _WIN32
    DWORD n = sizeof(buf);
    if (GetComputerNameA(buf, &n)) return lum_dup(buf);
    return lum_dup("");
#else
    if (gethostname(buf, sizeof(buf)) == 0) { buf[sizeof(buf) - 1] = 0; return lum_dup(buf); }
    return lum_dup("");
#endif
}

const char *lum_os_temp_dir(void) {
    const char *t = getenv("TEMP");
    if (!t) t = getenv("TMP");
    if (!t) t = getenv("TMPDIR");
#ifdef _WIN32
    if (!t) t = "C:\\Windows\\Temp";
#else
    if (!t) t = "/tmp";
#endif
    return lum_dup(t);
}

const char *lum_os_username(void) {
    const char *u = getenv("USERNAME");
    if (!u) u = getenv("USER");
    return lum_dup(u ? u : "");
}

int64_t lum_os_arch(void) {
#if defined(_WIN64) || defined(__x86_64__) || defined(__aarch64__)
    return 64;
#else
    return 32;
#endif
}
