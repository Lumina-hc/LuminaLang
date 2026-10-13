#include "lumina.h"

/* ---- shared helpers ---- */

static uint32_t lum_rotl32(uint32_t x, int c) { return (x << c) | (x >> (32 - c)); }
static uint32_t lum_rotr32(uint32_t x, int c) { return (x >> c) | (x << (32 - c)); }

static const char *lum_hex(const uint8_t *d, size_t n) {
    static const char *hx = "0123456789abcdef";
    char *r = (char *)lum_alloc(n * 2 + 1);
    for (size_t i = 0; i < n; i++) {
        r[i * 2] = hx[d[i] >> 4];
        r[i * 2 + 1] = hx[d[i] & 15];
    }
    r[n * 2] = 0;
    return r;
}

/* ---- MD5 (RFC 1321) ---- */

static void lum_md5_block(uint32_t state[4], const uint8_t block[64]) {
    static const uint32_t K[64] = {
        0xd76aa478, 0xe8c7b756, 0x242070db, 0xc1bdceee, 0xf57c0faf, 0x4787c62a,
        0xa8304613, 0xfd469501, 0x698098d8, 0x8b44f7af, 0xffff5bb1, 0x895cd7be,
        0x6b901122, 0xfd987193, 0xa679438e, 0x49b40821, 0xf61e2562, 0xc040b340,
        0x265e5a51, 0xe9b6c7aa, 0xd62f105d, 0x02441453, 0xd8a1e681, 0xe7d3fbc8,
        0x21e1cde6, 0xc33707d6, 0xf4d50d87, 0x455a14ed, 0xa9e3e905, 0xfcefa3f8,
        0x676f02d9, 0x8d2a4c8a, 0xfffa3942, 0x8771f681, 0x6d9d6122, 0xfde5380c,
        0xa4beea44, 0x4bdecfa9, 0xf6bb4b60, 0xbebfbc70, 0x289b7ec6, 0xeaa127fa,
        0xd4ef3085, 0x04881d05, 0xd9d4d039, 0xe6db99e5, 0x1fa27cf8, 0xc4ac5665,
        0xf4292244, 0x432aff97, 0xab9423a7, 0xfc93a039, 0x655b59c3, 0x8f0ccc92,
        0xffeff47d, 0x85845dd1, 0x6fa87e4f, 0xfe2ce6e0, 0xa3014314, 0x4e0811a1,
        0xf7537e82, 0xbd3af235, 0x2ad7d2bb, 0xeb86d391
    };
    static const int S[64] = {
        7, 12, 17, 22, 7, 12, 17, 22, 7, 12, 17, 22, 7, 12, 17, 22,
        5, 9, 14, 20, 5, 9, 14, 20, 5, 9, 14, 20, 5, 9, 14, 20,
        4, 11, 16, 23, 4, 11, 16, 23, 4, 11, 16, 23, 4, 11, 16, 23,
        6, 10, 15, 21, 6, 10, 15, 21, 6, 10, 15, 21, 6, 10, 15, 21
    };
    uint32_t m[16];
    for (int i = 0; i < 16; i++) {
        m[i] = (uint32_t)block[i * 4] | ((uint32_t)block[i * 4 + 1] << 8) |
               ((uint32_t)block[i * 4 + 2] << 16) | ((uint32_t)block[i * 4 + 3] << 24);
    }
    uint32_t a = state[0], b = state[1], c = state[2], d = state[3];
    for (int i = 0; i < 64; i++) {
        uint32_t f;
        int g;
        if (i < 16) { f = (b & c) | (~b & d); g = i; }
        else if (i < 32) { f = (d & b) | (~d & c); g = (5 * i + 1) % 16; }
        else if (i < 48) { f = b ^ c ^ d; g = (3 * i + 5) % 16; }
        else { f = c ^ (b | ~d); g = (7 * i) % 16; }
        uint32_t tmp = d;
        d = c;
        c = b;
        b = b + lum_rotl32(a + f + K[i] + m[g], S[i]);
        a = tmp;
    }
    state[0] += a; state[1] += b; state[2] += c; state[3] += d;
}

const char *lum_enc_md5(const char *s) {
    const uint8_t *msg = (const uint8_t *)(s ? s : "");
    size_t len = strlen((const char *)msg);
    uint32_t state[4] = { 0x67452301, 0xefcdab89, 0x98badcfe, 0x10325476 };
    uint64_t bitlen = (uint64_t)len * 8;

    size_t i = 0;
    for (; i + 64 <= len; i += 64) lum_md5_block(state, msg + i);
    uint8_t tail[128];
    size_t rem = len - i;
    memcpy(tail, msg + i, rem);
    tail[rem] = 0x80;
    size_t tl = (rem + 1 <= 56) ? 64 : 128;
    memset(tail + rem + 1, 0, tl - rem - 1);
    for (int j = 0; j < 8; j++) tail[tl - 8 + j] = (uint8_t)(bitlen >> (8 * j));
    lum_md5_block(state, tail);
    if (tl == 128) lum_md5_block(state, tail + 64);

    uint8_t dig[16];
    for (int j = 0; j < 4; j++)
        for (int k = 0; k < 4; k++)
            dig[j * 4 + k] = (uint8_t)(state[j] >> (8 * k));
    return lum_hex(dig, 16);
}

/* ---- SHA-1 (RFC 3174) ---- */

static void lum_sha1_block(uint32_t h[5], const uint8_t block[64]) {
    uint32_t w[80];
    for (int i = 0; i < 16; i++)
        w[i] = ((uint32_t)block[i * 4] << 24) | ((uint32_t)block[i * 4 + 1] << 16) |
               ((uint32_t)block[i * 4 + 2] << 8) | (uint32_t)block[i * 4 + 3];
    for (int i = 16; i < 80; i++)
        w[i] = lum_rotl32(w[i - 3] ^ w[i - 8] ^ w[i - 14] ^ w[i - 16], 1);

    uint32_t a = h[0], b = h[1], c = h[2], d = h[3], e = h[4];
    for (int i = 0; i < 80; i++) {
        uint32_t f, k;
        if (i < 20) { f = (b & c) | (~b & d); k = 0x5A827999; }
        else if (i < 40) { f = b ^ c ^ d; k = 0x6ED9EBA1; }
        else if (i < 60) { f = (b & c) | (b & d) | (c & d); k = 0x8F1BBCDC; }
        else { f = b ^ c ^ d; k = 0xCA62C1D6; }
        uint32_t tmp = lum_rotl32(a, 5) + f + e + k + w[i];
        e = d; d = c; c = lum_rotl32(b, 30); b = a; a = tmp;
    }
    h[0] += a; h[1] += b; h[2] += c; h[3] += d; h[4] += e;
}

const char *lum_enc_sha1(const char *s) {
    const uint8_t *msg = (const uint8_t *)(s ? s : "");
    size_t len = strlen((const char *)msg);
    uint32_t h[5] = { 0x67452301, 0xEFCDAB89, 0x98BADCFE, 0x10325476, 0xC3D2E1F0 };
    uint64_t bitlen = (uint64_t)len * 8;

    size_t i = 0;
    for (; i + 64 <= len; i += 64) lum_sha1_block(h, msg + i);
    uint8_t tail[128];
    size_t rem = len - i;
    memcpy(tail, msg + i, rem);
    tail[rem] = 0x80;
    size_t tl = (rem + 1 <= 56) ? 64 : 128;
    memset(tail + rem + 1, 0, tl - rem - 1);
    for (int j = 0; j < 8; j++) tail[tl - 1 - j] = (uint8_t)(bitlen >> (8 * j));
    lum_sha1_block(h, tail);
    if (tl == 128) lum_sha1_block(h, tail + 64);

    uint8_t dig[20];
    for (int j = 0; j < 5; j++)
        for (int k = 0; k < 4; k++)
            dig[j * 4 + k] = (uint8_t)(h[j] >> (24 - 8 * k));
    return lum_hex(dig, 20);
}

/* ---- SHA-256 (FIPS 180-4) ---- */

static const uint32_t lum_sha256_K[64] = {
    0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1,
    0x923f82a4, 0xab1c5ed5, 0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3,
    0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174, 0xe49b69c1, 0xefbe4786,
    0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
    0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147,
    0x06ca6351, 0x14292967, 0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13,
    0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85, 0xa2bfe8a1, 0xa81a664b,
    0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
    0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa4a,
    0x5b9cca4f, 0x682e6ff3, 0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208,
    0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2
};

static void lum_sha256_block(uint32_t h[8], const uint8_t block[64]) {
    uint32_t w[64];
    for (int i = 0; i < 16; i++)
        w[i] = ((uint32_t)block[i * 4] << 24) | ((uint32_t)block[i * 4 + 1] << 16) |
               ((uint32_t)block[i * 4 + 2] << 8) | (uint32_t)block[i * 4 + 3];
    for (int i = 16; i < 64; i++) {
        uint32_t s0 = lum_rotr32(w[i - 15], 7) ^ lum_rotr32(w[i - 15], 18) ^ (w[i - 15] >> 3);
        uint32_t s1 = lum_rotr32(w[i - 2], 17) ^ lum_rotr32(w[i - 2], 19) ^ (w[i - 2] >> 10);
        w[i] = w[i - 16] + s0 + w[i - 7] + s1;
    }
    uint32_t a = h[0], b = h[1], c = h[2], d = h[3];
    uint32_t e = h[4], f = h[5], g = h[6], hh = h[7];
    for (int i = 0; i < 64; i++) {
        uint32_t S1 = lum_rotr32(e, 6) ^ lum_rotr32(e, 11) ^ lum_rotr32(e, 25);
        uint32_t ch = (e & f) ^ (~e & g);
        uint32_t t1 = hh + S1 + ch + lum_sha256_K[i] + w[i];
        uint32_t S0 = lum_rotr32(a, 2) ^ lum_rotr32(a, 13) ^ lum_rotr32(a, 22);
        uint32_t maj = (a & b) ^ (a & c) ^ (b & c);
        uint32_t t2 = S0 + maj;
        hh = g; g = f; f = e; e = d + t1;
        d = c; c = b; b = a; a = t1 + t2;
    }
    h[0] += a; h[1] += b; h[2] += c; h[3] += d;
    h[4] += e; h[5] += f; h[6] += g; h[7] += hh;
}

const char *lum_enc_sha256(const char *s) {
    const uint8_t *msg = (const uint8_t *)(s ? s : "");
    size_t len = strlen((const char *)msg);
    uint32_t h[8] = {
        0x6a09e667, 0xbb67ae85, 0x3c6ef372, 0xa54ff53a,
        0x510e527f, 0x9b05688c, 0x1f83d9ab, 0x5be0cd19
    };
    uint64_t bitlen = (uint64_t)len * 8;

    size_t i = 0;
    for (; i + 64 <= len; i += 64) lum_sha256_block(h, msg + i);
    uint8_t tail[128];
    size_t rem = len - i;
    memcpy(tail, msg + i, rem);
    tail[rem] = 0x80;
    size_t tl = (rem + 1 <= 56) ? 64 : 128;
    memset(tail + rem + 1, 0, tl - rem - 1);
    for (int j = 0; j < 8; j++) tail[tl - 1 - j] = (uint8_t)(bitlen >> (8 * j));
    lum_sha256_block(h, tail);
    if (tl == 128) lum_sha256_block(h, tail + 64);

    uint8_t dig[32];
    for (int j = 0; j < 8; j++)
        for (int k = 0; k < 4; k++)
            dig[j * 4 + k] = (uint8_t)(h[j] >> (24 - 8 * k));
    return lum_hex(dig, 32);
}

/* ---- CRC-32 (IEEE 802.3) ---- */

int64_t lum_enc_crc32(const char *s) {
    const unsigned char *p = (const unsigned char *)(s ? s : "");
    uint32_t crc = 0xFFFFFFFFu;
    while (*p) {
        crc ^= (uint32_t)*p++;
        for (int i = 0; i < 8; i++)
            crc = (crc >> 1) ^ (0xEDB88320u & (uint32_t)(-(int32_t)(crc & 1)));
    }
    return (int64_t)(crc ^ 0xFFFFFFFFu);
}

/* ---- Base64 ---- */

static const char lum_b64_alpha[] =
    "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";

const char *lum_enc_base64_encode(const char *s) {
    const unsigned char *p = (const unsigned char *)(s ? s : "");
    size_t n = strlen((const char *)p);
    size_t outlen = ((n + 2) / 3) * 4;
    char *r = (char *)lum_alloc(outlen + 1);
    size_t i = 0, o = 0;
    while (i + 3 <= n) {
        uint32_t v = ((uint32_t)p[i] << 16) | ((uint32_t)p[i + 1] << 8) | p[i + 2];
        r[o++] = lum_b64_alpha[(v >> 18) & 63];
        r[o++] = lum_b64_alpha[(v >> 12) & 63];
        r[o++] = lum_b64_alpha[(v >> 6) & 63];
        r[o++] = lum_b64_alpha[v & 63];
        i += 3;
    }
    size_t rem = n - i;
    if (rem == 1) {
        uint32_t v = (uint32_t)p[i] << 16;
        r[o++] = lum_b64_alpha[(v >> 18) & 63];
        r[o++] = lum_b64_alpha[(v >> 12) & 63];
        r[o++] = '=';
        r[o++] = '=';
    } else if (rem == 2) {
        uint32_t v = ((uint32_t)p[i] << 16) | ((uint32_t)p[i + 1] << 8);
        r[o++] = lum_b64_alpha[(v >> 18) & 63];
        r[o++] = lum_b64_alpha[(v >> 12) & 63];
        r[o++] = lum_b64_alpha[(v >> 6) & 63];
        r[o++] = '=';
    }
    r[o] = 0;
    return r;
}

static int lum_b64_val(int c) {
    if (c >= 'A' && c <= 'Z') return c - 'A';
    if (c >= 'a' && c <= 'z') return c - 'a' + 26;
    if (c >= '0' && c <= '9') return c - '0' + 52;
    if (c == '+') return 62;
    if (c == '/') return 63;
    return -1;
}

const char *lum_enc_base64_decode(const char *s) {
    const unsigned char *p = (const unsigned char *)(s ? s : "");
    size_t n = strlen((const char *)p);
    char *r = (char *)lum_alloc((n / 4) * 3 + 1);
    size_t o = 0;
    int quad[4], qn = 0;
    for (size_t i = 0; i < n; i++) {
        int c = p[i];
        if (c == '=') break;
        int v = lum_b64_val(c);
        if (v < 0) continue;
        quad[qn++] = v;
        if (qn == 4) {
            uint32_t x = ((uint32_t)quad[0] << 18) | ((uint32_t)quad[1] << 12) |
                         ((uint32_t)quad[2] << 6) | (uint32_t)quad[3];
            r[o++] = (char)((x >> 16) & 0xFF);
            r[o++] = (char)((x >> 8) & 0xFF);
            r[o++] = (char)(x & 0xFF);
            qn = 0;
        }
    }
    if (qn == 2) {
        uint32_t x = ((uint32_t)quad[0] << 18) | ((uint32_t)quad[1] << 12);
        r[o++] = (char)((x >> 16) & 0xFF);
    } else if (qn == 3) {
        uint32_t x = ((uint32_t)quad[0] << 18) | ((uint32_t)quad[1] << 12) |
                     ((uint32_t)quad[2] << 6);
        r[o++] = (char)((x >> 16) & 0xFF);
        r[o++] = (char)((x >> 8) & 0xFF);
    }
    r[o] = 0;
    return r;
}
