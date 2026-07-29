UNAME_S := $(shell uname -s)

ifeq ($(UNAME_S),Linux)
    TARGET = libruntime.so
    CFLAGS = -shared -fPIC -O2
endif
ifeq ($(UNAME_S),Darwin)
    TARGET = libruntime.dylib
    CFLAGS = -shared -fPIC -O2
endif
ifeq ($(OS),Windows_NT)
    TARGET = libruntime.dll
    CFLAGS = -shared -O2
endif

all: $(TARGET)

$(TARGET): runtime.c
	$(CC) $(CFLAGS) -o $(TARGET) runtime.c

clean:
	rm -f $(TARGET) *.o

.PHONY: all clean
