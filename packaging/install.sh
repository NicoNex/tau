#!/bin/sh
# Install tau from the archive this script came in.
#
# There is no build here and nothing is fetched: the archive already holds the
# interpreter, the runtime a bundled program is built on, and the standard
# library, and this only puts them somewhere and says where that was.
#
#	./install.sh                     into /usr/local as root, ~/.local otherwise
#	./install.sh --prefix=/opt/tau   somewhere else
#	./install.sh --uninstall         take it away again
#
# The interpreter looks for its library beside itself, so any prefix works and
# none of them needs a variable set afterwards.

set -eu

usage() {
	cat <<'EOF'
Usage: ./install.sh [--prefix=DIR] [--uninstall]

  --prefix=DIR  where to install; default /usr/local as root, ~/.local if not
  --uninstall   remove what a previous run of this installed
  -h, --help    this

The binary goes in PREFIX/bin and the standard library in PREFIX/lib/tau.
EOF
}

prefix=
uninstall=no

for arg in "$@"; do
	case $arg in
	--prefix=*) prefix=${arg#--prefix=} ;;
	--uninstall) uninstall=yes ;;
	-h | --help)
		usage
		exit 0
		;;
	*)
		echo "install.sh: $arg is not an option I know" >&2
		usage >&2
		exit 2
		;;
	esac
done

# Where this script is, which is where the tree it installs is too. Running it
# from somewhere else has to work, so nothing here is relative to the caller's
# directory.
here=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)

if [ -z "$prefix" ]; then
	if [ "$(id -u)" = 0 ]; then
		prefix=/usr/local
	else
		prefix=$HOME/.local
	fi
fi

# An empty or relative prefix would make the paths below mean something else
# entirely, and one of them is deleted.
case $prefix in
/*) ;;
*)
	echo "install.sh: the prefix has to be an absolute path, got $prefix" >&2
	exit 2
	;;
esac

bin=$prefix/bin/tau
lib=$prefix/lib/tau

if [ "$uninstall" = yes ]; then
	rm -f "$bin"
	# Only ever the directory this puts there, never a prefix.
	case $lib in
	*/lib/tau) rm -rf "$lib" ;;
	*)
		echo "install.sh: refusing to delete $lib" >&2
		exit 1
		;;
	esac
	echo "removed $bin"
	echo "removed $lib"
	exit 0
fi

if [ ! -f "$here/bin/tau" ] || [ ! -d "$here/lib/tau" ]; then
	echo "install.sh: this does not look like the tau archive: no bin/tau and lib/tau next to me" >&2
	exit 1
fi

# The archive is built for one machine. Saying so now is better than a
# confusing failure at the first run.
want=$(cat "$here/lib/tau/ARCH" 2>/dev/null || echo unknown)
have=$(uname -m 2>/dev/null || echo unknown)
if [ "$want" != unknown ] && [ "$have" != unknown ] && [ "$want" != "$have" ]; then
	echo "install.sh: this archive is for $want and this machine is $have" >&2
	exit 1
fi

if ! mkdir -p "$prefix/bin" "$prefix/lib" 2>/dev/null; then
	echo "install.sh: cannot write to $prefix - run with sudo, or pass --prefix=\$HOME/.local" >&2
	exit 1
fi

# The library goes first and whole, so that a module dropped from the standard
# library does not stay installed forever.
case $lib in
*/lib/tau) rm -rf "$lib" ;;
esac

cp -R "$here/lib/tau" "$lib"
cp "$here/bin/tau" "$bin"
chmod 755 "$bin"

echo
echo "tau        $bin"
echo "stdlib     $lib"

case ":$PATH:" in
*":$prefix/bin:"*) ;;
*) echo "note: $prefix/bin is not in your PATH" ;;
esac

echo
"$bin" version
