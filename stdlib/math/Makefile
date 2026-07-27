UNAME_S := $(shell uname -s)

ifeq ($(UNAME_S),Linux)
    TARGET = libmath.so
    CFLAGS = -shared -fPIC -O2
endif
ifeq ($(UNAME_S),Darwin)
    TARGET = libmath.dylib
    CFLAGS = -shared -fPIC -O2
endif
ifeq ($(OS),Windows_NT)
    TARGET = libmath.dll
    CFLAGS = -shared -O2
endif

all: $(TARGET)

$(TARGET): math.c
	$(CC) $(CFLAGS) -o $(TARGET) math.c

clean:
	rm -f $(TARGET) *.o

.PHONY: all clean
