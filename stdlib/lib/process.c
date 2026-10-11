#include "lumina.h"

int64_t lum_proc_system(const char *cmd) {
    if (!cmd) return -1;
    return (int64_t)system(cmd);
}
