#include "lumina.h"
#include <time.h>

static int lum_rand_seeded = 0;

static void lum_rand_ensure(void) {
    if (!lum_rand_seeded) {
        srand((unsigned)time(NULL));
        lum_rand_seeded = 1;
    }
}

int64_t lum_rand_seed(int64_t seed) {
    srand((unsigned)seed);
    lum_rand_seeded = 1;
    return 0;
}

int64_t lum_rand_int(int64_t lo, int64_t hi) {
    lum_rand_ensure();
    if (hi <= lo) return lo;
    double span = (double)(hi - lo);
    return lo + (int64_t)((double)rand() / ((double)RAND_MAX + 1.0) * span);
}

double lum_rand_f64(void) {
    lum_rand_ensure();
    return (double)rand() / ((double)RAND_MAX + 1.0);
}
