#include "lumina.h"
#include <time.h>

#ifdef _WIN32
#include <windows.h>
#else
#include <sys/time.h>
#endif

int64_t lum_time_now(void) { return (int64_t)time(NULL); }

int64_t lum_time_now_ms(void) {
#ifdef _WIN32
    FILETIME ft;
    GetSystemTimeAsFileTime(&ft);
    ULARGE_INTEGER u;
    u.LowPart = ft.dwLowDateTime;
    u.HighPart = ft.dwHighDateTime;
    return (int64_t)((u.QuadPart - 116444736000000000ULL) / 10000ULL);
#else
    struct timeval tv;
    gettimeofday(&tv, NULL);
    return (int64_t)tv.tv_sec * 1000 + tv.tv_usec / 1000;
#endif
}

static void lum_localtime(struct tm *out) {
    time_t t = time(NULL);
#ifdef _WIN32
    localtime_s(out, &t);
#else
    localtime_r(&t, out);
#endif
}

int64_t lum_time_year(void)    { struct tm t; lum_localtime(&t); return t.tm_year + 1900; }
int64_t lum_time_month(void)   { struct tm t; lum_localtime(&t); return t.tm_mon + 1; }
int64_t lum_time_day(void)     { struct tm t; lum_localtime(&t); return t.tm_mday; }
int64_t lum_time_hour(void)    { struct tm t; lum_localtime(&t); return t.tm_hour; }
int64_t lum_time_minute(void)  { struct tm t; lum_localtime(&t); return t.tm_min; }
int64_t lum_time_second(void)  { struct tm t; lum_localtime(&t); return t.tm_sec; }
int64_t lum_time_weekday(void) { struct tm t; lum_localtime(&t); return t.tm_wday; }
int64_t lum_time_yearday(void) { struct tm t; lum_localtime(&t); return t.tm_yday + 1; }

int64_t lum_time_sleep_ms(int64_t ms) {
    if (ms < 0) ms = 0;
#ifdef _WIN32
    Sleep((DWORD)ms);
#else
    struct timespec ts;
    ts.tv_sec = ms / 1000;
    ts.tv_nsec = (ms % 1000) * 1000000L;
    nanosleep(&ts, NULL);
#endif
    return 0;
}

double lum_time_unix_f(void) {
    return (double)lum_time_now_ms() / 1000.0;
}

const char *lum_time_now_str(void) {
    time_t t = time(NULL);
    struct tm tm;
    memset(&tm, 0, sizeof(tm));
#ifdef _WIN32
    localtime_s(&tm, &t);
#else
    localtime_r(&t, &tm);
#endif
    char buf[40];
    if (strftime(buf, sizeof(buf), "%Y-%m-%d %H:%M:%S", &tm) == 0) {
        buf[0] = 0;
    }
    return lum_dup(buf);
}

int64_t lum_time_utc_offset(void) {
    time_t t = time(NULL);
    struct tm lt, gt;
    memset(&lt, 0, sizeof(lt));
    memset(&gt, 0, sizeof(gt));
#ifdef _WIN32
    localtime_s(&lt, &t);
    gmtime_s(&gt, &t);
#else
    localtime_r(&t, &lt);
    gmtime_r(&t, &gt);
#endif
    time_t local = mktime(&lt);
    time_t utc = mktime(&gt);
    return (int64_t)difftime(local, utc);
}
