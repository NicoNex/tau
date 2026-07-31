UNAME_S := $(shell uname -s)

ifeq ($(UNAME_S),Linux)
    TARGET = libsyncatomic.so
    CFLAGS = -shared -fPIC -O2
endif
ifeq ($(UNAME_S),Darwin)
    TARGET = libsyncatomic.dylib
    CFLAGS = -shared -fPIC -O2
endif
ifeq ($(OS),Windows_NT)
    TARGET = libsyncatomic.dll
    CFLAGS = -shared -O2
endif

all: $(TARGET)

$(TARGET): atomic.c
	$(CC) $(CFLAGS) -o $(TARGET) atomic.c

clean:
	rm -f $(TARGET) *.o

.PHONY: all clean
