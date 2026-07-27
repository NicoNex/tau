#include <stdint.h>
#include <string.h>
#include <time.h>
#include <unistd.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <errno.h>

#if !defined(_WIN32) && !defined(WIN32)
    #include <sys/socket.h>
    #include <netinet/in.h>
    #include <arpa/inet.h>
    #include <netdb.h>
#else
    #include <winsock2.h>
    #include <ws2tcpip.h>
    #pragma comment(lib, "ws2_32.lib")
    #define SHUT_RD SD_RECEIVE
    #define SHUT_WR SD_SEND
    #define SHUT_RDWR SD_BOTH
#endif

// ========== File Operations ==========

int64_t sys_open(const char *path, int64_t flags, int64_t mode) {
#if !defined(_WIN32) && !defined(WIN32)
    return open(path, (int)flags, (mode_t)mode);
#else
    // Windows uses _open with different flags
    int win_flags = 0;
    if (flags & 0x01) win_flags |= _O_RDONLY;
    if (flags & 0x02) win_flags |= _O_WRONLY;
    if (flags & 0x100) win_flags |= _O_CREAT;
    if (flags & 0x200) win_flags |= _O_TRUNC;
    if (flags & 0x400) win_flags |= _O_APPEND;
    return _open(path, win_flags, (int)mode);
#endif
}

int64_t sys_close(int64_t fd) {
#if !defined(_WIN32) && !defined(WIN32)
    return close((int)fd);
#else
    return _close((int)fd);
#endif
}

int64_t sys_read(int64_t fd, void *buf, int64_t count) {
#if !defined(_WIN32) && !defined(WIN32)
    return read((int)fd, buf, (size_t)count);
#else
    return _read((int)fd, buf, (unsigned int)count);
#endif
}

int64_t sys_write(int64_t fd, const void *buf, int64_t count) {
#if !defined(_WIN32) && !defined(WIN32)
    return write((int)fd, buf, (size_t)count);
#else
    return _write((int)fd, buf, (unsigned int)count);
#endif
}

int64_t sys_lseek(int64_t fd, int64_t offset, int64_t whence) {
#if !defined(_WIN32) && !defined(WIN32)
    return lseek((int)fd, (off_t)offset, (int)whence);
#else
    return _lseek((int)fd, (long)offset, (int)whence);
#endif
}

int64_t sys_unlink(const char *path) {
#if !defined(_WIN32) && !defined(WIN32)
    return unlink(path);
#else
    return _unlink(path);
#endif
}

int64_t sys_mkdir(const char *path, int64_t mode) {
#if !defined(_WIN32) && !defined(WIN32)
    return mkdir(path, (mode_t)mode);
#else
    return _mkdir(path);
#endif
}

int64_t sys_rmdir(const char *path) {
#if !defined(_WIN32) && !defined(WIN32)
    return rmdir(path);
#else
    return _rmdir(path);
#endif
}

int64_t sys_chmod(const char *path, int64_t mode) {
#if !defined(_WIN32) && !defined(WIN32)
    return chmod(path, (mode_t)mode);
#else
    return _chmod(path, (int)mode);
#endif
}

int64_t sys_access(const char *path, int64_t mode) {
#if !defined(_WIN32) && !defined(WIN32)
    return access(path, (int)mode);
#else
    return _access(path, (int)mode);
#endif
}

// ========== Network Operations ==========

#if defined(_WIN32) || defined(WIN32)
static int winsock_initialized = 0;

static void ensure_winsock() {
    if (!winsock_initialized) {
        WSADATA wsa;
        WSAStartup(MAKEWORD(2, 2), &wsa);
        winsock_initialized = 1;
    }
}
#endif

int64_t sys_socket(int64_t domain, int64_t type, int64_t protocol) {
#if !defined(_WIN32) && !defined(WIN32)
    return socket((int)domain, (int)type, (int)protocol);
#else
    ensure_winsock();
    return (int64_t)socket((int)domain, (int)type, (int)protocol);
#endif
}

int64_t sys_bind(int64_t sockfd, const void *addr, int64_t addrlen) {
#if !defined(_WIN32) && !defined(WIN32)
    return bind((int)sockfd, (const struct sockaddr *)addr, (socklen_t)addrlen);
#else
    return bind((SOCKET)sockfd, (const struct sockaddr *)addr, (int)addrlen);
#endif
}

int64_t sys_listen(int64_t sockfd, int64_t backlog) {
#if !defined(_WIN32) && !defined(WIN32)
    return listen((int)sockfd, (int)backlog);
#else
    return listen((SOCKET)sockfd, (int)backlog);
#endif
}

int64_t sys_accept(int64_t sockfd, void *addr, void *addrlen) {
#if !defined(_WIN32) && !defined(WIN32)
    return accept((int)sockfd, (struct sockaddr *)addr, (socklen_t *)addrlen);
#else
    return (int64_t)accept((SOCKET)sockfd, (struct sockaddr *)addr, (int *)addrlen);
#endif
}

int64_t sys_connect(int64_t sockfd, const void *addr, int64_t addrlen) {
#if !defined(_WIN32) && !defined(WIN32)
    return connect((int)sockfd, (const struct sockaddr *)addr, (socklen_t)addrlen);
#else
    return connect((SOCKET)sockfd, (const struct sockaddr *)addr, (int)addrlen);
#endif
}

int64_t sys_send(int64_t sockfd, const void *buf, int64_t len, int64_t flags) {
#if !defined(_WIN32) && !defined(WIN32)
    return send((int)sockfd, buf, (size_t)len, (int)flags);
#else
    return send((SOCKET)sockfd, (const char *)buf, (int)len, (int)flags);
#endif
}

int64_t sys_recv(int64_t sockfd, void *buf, int64_t len, int64_t flags) {
#if !defined(_WIN32) && !defined(WIN32)
    return recv((int)sockfd, buf, (size_t)len, (int)flags);
#else
    return recv((SOCKET)sockfd, (char *)buf, (int)len, (int)flags);
#endif
}

int64_t sys_sendto(int64_t sockfd, const void *buf, int64_t len, int64_t flags,
                   const void *dest_addr, int64_t addrlen) {
#if !defined(_WIN32) && !defined(WIN32)
    return sendto((int)sockfd, buf, (size_t)len, (int)flags,
                  (const struct sockaddr *)dest_addr, (socklen_t)addrlen);
#else
    return sendto((SOCKET)sockfd, (const char *)buf, (int)len, (int)flags,
                  (const struct sockaddr *)dest_addr, (int)addrlen);
#endif
}

int64_t sys_recvfrom(int64_t sockfd, void *buf, int64_t len, int64_t flags,
                     void *src_addr, void *addrlen) {
#if !defined(_WIN32) && !defined(WIN32)
    return recvfrom((int)sockfd, buf, (size_t)len, (int)flags,
                    (struct sockaddr *)src_addr, (socklen_t *)addrlen);
#else
    return recvfrom((SOCKET)sockfd, (char *)buf, (int)len, (int)flags,
                    (struct sockaddr *)src_addr, (int *)addrlen);
#endif
}

int64_t sys_setsockopt(int64_t sockfd, int64_t level, int64_t optname,
                       const void *optval, int64_t optlen) {
#if !defined(_WIN32) && !defined(WIN32)
    return setsockopt((int)sockfd, (int)level, (int)optname,
                      optval, (socklen_t)optlen);
#else
    return setsockopt((SOCKET)sockfd, (int)level, (int)optname,
                      (const char *)optval, (int)optlen);
#endif
}

int64_t sys_getsockopt(int64_t sockfd, int64_t level, int64_t optname,
                       void *optval, void *optlen) {
#if !defined(_WIN32) && !defined(WIN32)
    return getsockopt((int)sockfd, (int)level, (int)optname,
                      optval, (socklen_t *)optlen);
#else
    return getsockopt((SOCKET)sockfd, (int)level, (int)optname,
                      (char *)optval, (int *)optlen);
#endif
}

int64_t sys_shutdown(int64_t sockfd, int64_t how) {
#if !defined(_WIN32) && !defined(WIN32)
    return shutdown((int)sockfd, (int)how);
#else
    return shutdown((SOCKET)sockfd, (int)how);
#endif
}

// ========== Network Helpers ==========

uint16_t sys_htons(uint16_t hostshort) {
    return htons(hostshort);
}

uint16_t sys_ntohs(uint16_t netshort) {
    return ntohs(netshort);
}

uint32_t sys_htonl(uint32_t hostlong) {
    return htonl(hostlong);
}

uint32_t sys_ntohl(uint32_t netlong) {
    return ntohl(netlong);
}

uint32_t sys_inet_addr(const char *cp) {
    return inet_addr(cp);
}

// ========== Address Helpers ==========

// Fills buf with a sockaddr_in for ip:port and returns its size, so that the
// callers don't have to lay out the struct byte by byte.
int64_t sys_sockaddr_in(void *buf, const char *ip, int64_t port) {
    struct sockaddr_in addr;

    memset(&addr, 0, sizeof(addr));
    addr.sin_family = AF_INET;
    addr.sin_port = htons((uint16_t)port);
    addr.sin_addr.s_addr = (ip == NULL || *ip == '\0') ? INADDR_ANY : inet_addr(ip);

    memcpy(buf, &addr, sizeof(addr));
    return sizeof(addr);
}

int64_t sys_sockaddr_in_size() {
    return sizeof(struct sockaddr_in);
}

// Port of a sockaddr_in previously filled by accept or recvfrom.
int64_t sys_sockaddr_port(const void *buf) {
    return ntohs(((const struct sockaddr_in *)buf)->sin_port);
}

// Writes the dotted address of a sockaddr_in into out, returns its length.
int64_t sys_sockaddr_ip(const void *buf, void *out, int64_t outlen) {
    const char *ip = inet_ntoa(((const struct sockaddr_in *)buf)->sin_addr);
    size_t len = strlen(ip);

    if ((int64_t)len >= outlen) len = outlen > 0 ? outlen - 1 : 0;
    memcpy(out, ip, len);
    ((char *)out)[len] = '\0';

    return len;
}

// Writes a native int32 into buf, for the calls that want a pointer to one
// (the addrlen of accept, the value of setsockopt).
int64_t sys_int32(void *buf, int64_t v) {
    int32_t n = (int32_t)v;

    memcpy(buf, &n, sizeof(n));
    return sizeof(n);
}

int64_t sys_int32_size() {
    return sizeof(int32_t);
}

// Writes a struct timeval of ms milliseconds into buf.
int64_t sys_timeval(void *buf, int64_t ms) {
    struct timeval tv;

    tv.tv_sec = ms / 1000;
    tv.tv_usec = (ms % 1000) * 1000;
    memcpy(buf, &tv, sizeof(tv));

    return sizeof(tv);
}

int64_t sys_timeval_size() {
    return sizeof(struct timeval);
}

// ========== Time ==========

int64_t sys_time_unix() {
    return (int64_t)time(NULL);
}

int64_t sys_time_ms() {
    struct timespec ts;

    if (clock_gettime(CLOCK_REALTIME, &ts) != 0) return -1;
    return (int64_t)ts.tv_sec * 1000 + ts.tv_nsec / 1000000;
}

// Monotonic clock: unaffected by changes of the wall clock, for measuring.
int64_t sys_time_mono_ms() {
    struct timespec ts;

    if (clock_gettime(CLOCK_MONOTONIC, &ts) != 0) return -1;
    return (int64_t)ts.tv_sec * 1000 + ts.tv_nsec / 1000000;
}

int64_t sys_sleep_ms(int64_t ms) {
    if (ms <= 0) return 0;

#if !defined(_WIN32) && !defined(WIN32)
    struct timespec ts = {
        .tv_sec = ms / 1000,
        .tv_nsec = (ms % 1000) * 1000000
    };
    // Finishes the wait even if a signal interrupts it.
    while (nanosleep(&ts, &ts) == -1 && errno == EINTR) {}
    return 0;
#else
    Sleep((DWORD)ms);
    return 0;
#endif
}

// ========== Error Handling ==========

// Writes the message of an errno into out, returns its length.
int64_t sys_strerror(int64_t err, void *out, int64_t outlen) {
    const char *msg = strerror((int)err);
    size_t len = strlen(msg);

    if ((int64_t)len >= outlen) len = outlen > 0 ? outlen - 1 : 0;
    memcpy(out, msg, len);
    ((char *)out)[len] = '\0';

    return len;
}

int64_t sys_errno() {
    return errno;
}

void sys_set_errno(int64_t err) {
    errno = (int)err;
}
