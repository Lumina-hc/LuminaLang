#include "lumina.h"

#ifdef _WIN32
#include <windows.h>

int64_t lum_win_msgbox(const char *text, const char *title) {
    return (int64_t)MessageBoxA(NULL, text ? text : "", title ? title : "LuminaLang",
                                MB_OK | MB_ICONINFORMATION);
}

int64_t lum_win_beep(int64_t freq, int64_t ms) {
    return Beep((DWORD)freq, (DWORD)ms) ? 1 : 0;
}

int64_t lum_win_tick_ms(void) { return (int64_t)GetTickCount64(); }

int64_t lum_win_console_title(const char *title) {
    return SetConsoleTitleA(title ? title : "") ? 1 : 0;
}

int64_t lum_win_last_error(void) { return (int64_t)GetLastError(); }

int64_t lum_win_sleep_ms(int64_t ms) { Sleep((DWORD)(ms < 0 ? 0 : ms)); return 0; }

#else

int64_t lum_win_msgbox(const char *text, const char *title) { (void)text; (void)title; return 0; }
int64_t lum_win_beep(int64_t freq, int64_t ms) { (void)freq; (void)ms; return 0; }
int64_t lum_win_tick_ms(void) { return 0; }
int64_t lum_win_console_title(const char *title) { (void)title; return 0; }
int64_t lum_win_last_error(void) { return 0; }
int64_t lum_win_sleep_ms(int64_t ms) { (void)ms; return 0; }

#endif
