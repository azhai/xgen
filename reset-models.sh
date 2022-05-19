#!/bin/bash

find models/ -name "*.go" | grep -v mixin | xargs rm -f
rmdir --ignore-fail-on-non-empty models/*/

