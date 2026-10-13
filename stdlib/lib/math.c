#include "lumina.h"
#include <math.h>

double lum_math_sin(double x)   { return sin(x); }
double lum_math_cos(double x)   { return cos(x); }
double lum_math_tan(double x)   { return tan(x); }
double lum_math_asin(double x)  { return asin(x); }
double lum_math_acos(double x)  { return acos(x); }
double lum_math_atan(double x)  { return atan(x); }
double lum_math_atan2(double y, double x) { return atan2(y, x); }
double lum_math_sqrt(double x)  { return sqrt(x); }
double lum_math_pow(double x, double y)   { return pow(x, y); }
double lum_math_fabs(double x)  { return fabs(x); }
double lum_math_ln(double x)    { return log(x); }
double lum_math_log10(double x) { return log10(x); }
double lum_math_log2(double x)  { return log2(x); }
double lum_math_exp(double x)   { return exp(x); }
double lum_math_floor(double x) { return floor(x); }
double lum_math_ceil(double x)  { return ceil(x); }
double lum_math_round(double x) { return round(x); }
double lum_math_trunc(double x) { return trunc(x); }
double lum_math_fmod(double x, double y)  { return fmod(x, y); }
double lum_math_hypot(double x, double y) { return hypot(x, y); }

double lum_math_pi(void)  { return 3.14159265358979323846; }
double lum_math_e(void)   { return 2.71828182845904523536; }
double lum_math_tau(void) { return 6.28318530717958647692; }
