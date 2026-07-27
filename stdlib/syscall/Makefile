UNAME_S := $(shell uname -s)

ifeq ($(UNAME_S),Linux)
    TARGET = libsyscall.so
    CFLAGS = -shared -fPIC -O2
endif
ifeq ($(UNAME_S),Darwin)
    TARGET = libsyscall.dylib
    CFLAGS = -shared -fPIC -O2
endif
ifeq ($(OS),Windows_NT)
    TARGET = libsyscall.dll
    CFLAGS = -shared -O2 -lws2_32
endif

all: $(TARGET)

$(TARGET): syscall.c
	$(CC) $(CFLAGS) -o $(TARGET) syscall.c

clean:
	rm -f $(TARGET) *.o

.PHONY: all clean
