CC = clang
CXX = clang++

# Normal
CFLAGS := $(CFLAGS) -g -O3 -Wall -Wextra -pedantic -Werror -std=c18 -pthread
CXXFLAGS := $(CXXFLAGS) -g -O3 -Wall -Wextra -pedantic -Werror -std=c++20 -pthread


## To use Thread Sanitizer, ran with -O2 flag for reasonable performance
# CXXFLAGS := $(CXXFLAGS) -fsanitize=thread -g -O2 -Wall -Wextra -pedantic -Werror -std=c++20 -pthread
# CFLAGS := $(CFLAGS) -fsanitize=thread -g -O2 -Wall -Wextra -pedantic -Werror -std=c18 -pthread

## To use Address Sanitizer, ran with -O1
# CXXFLAGS := $(CXXFLAGS) -fsanitize=address -g -O1 -Wall -Wextra -pedantic -Werror -std=c++20
# CFLAGS := $(CFLAGS) -fsanitize=address -g -O1 -Wall -Wextra -pedantic -Werror -std=c18


# ## NOTE: Valgrind/Helgrind currently does not work
# ## To use Valgrind/Helgrind, we need to use these flags, or else it will give an error as valgrind 3.18.1 cannot read -gdwarf-5 
# ## files, which is being used in the latest version
# CFLAGS := $(CFLAGS) -gdwarf-4 -O1 -Wall -Wextra -pedantic -Werror -std=c18 -pthread
# CXXFLAGS := $(CXXFLAGS) -gdwarf-4 -O1 -Wall -Wextra -pedantic -Werror -std=c++20 -pthread



BUILDDIR = build

SRCS = main.cpp engine.cpp io.cpp

all: engine client

engine: $(SRCS:%=$(BUILDDIR)/%.o)
	$(LINK.cc) $^ $(LOADLIBES) $(LDLIBS) -o $@

client: $(BUILDDIR)/client.cpp.o
	$(LINK.cc) $^ $(LOADLIBES) $(LDLIBS) -o $@

.PHONY: clean
clean:
	rm -rf $(BUILDDIR)
	rm -f client engine

DEPFLAGS = -MT $@ -MMD -MP -MF $(BUILDDIR)/$<.d
COMPILE.cpp = $(CXX) $(DEPFLAGS) $(CXXFLAGS) $(CPPFLAGS) $(TARGET_ARCH) -c

$(BUILDDIR)/%.cpp.o: %.cpp | $(BUILDDIR)
	$(COMPILE.cpp) $(OUTPUT_OPTION) $<

$(BUILDDIR): ; @mkdir -p $@

DEPFILES := $(SRCS:%=$(BUILDDIR)/%.d) $(BUILDDIR)/client.cpp.d

-include $(DEPFILES)
