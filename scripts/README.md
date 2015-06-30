This is a collection of small utilities for bumping `pkgrel` or `pkgver` for PKGBUILD files.

Requirements
------------

* [setconf](https://github.com/xyproto/setconf)
* [getver](https://github.com/xyproto/getver)

Included scripts
----------------

* **bumpver** for updating the pkgver of the PKGBUILD in the same directory (or the file given as the first argument)
* **vup** for updating the pkgver and hash sums automatically. (Hash sums should ideally be provided by upstream).
* **geturlver** for retrieving the latest version for a PKGBUILD.
* **bumprel** uses `setconf` to increase the pkgrel number with 1.

If a line starting with "# getver: " is found, the rest of the line will be used as paramters for `getver`. This is useful for i.e. specifying the second found version number on the webpage, with `-u 2`. See the `getver` manpage for more information.
