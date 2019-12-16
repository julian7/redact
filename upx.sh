#!/bin/sh

/usr/bin/which -s upx || exit 0
upx dist/redact*/redact*
