UNAME_S := $(shell uname -s)

ifeq ($(UNAME_S),Linux)
    TARGET = libffitest.so
    CFLAGS = -shared -fPIC -O2
endif
ifeq ($(UNAME_S),Darwin)
    TARGET = libffitest.dylib
    CFLAGS = -shared -fPIC -O2
endif
ifeq ($(OS),Windows_NT)
    TARGET = libffitest.dll
    CFLAGS = -shared -O2
endif

all: $(TARGET)

$(TARGET): ffitest.c
	$(CC) $(CFLAGS) -o $(TARGET) ffitest.c

clean:
	rm -f $(TARGET) *.o

.PHONY: all clean
