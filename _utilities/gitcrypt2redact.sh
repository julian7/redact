#!/usr/bin/env bash

# 0a. git-crypt unlock
# 0b. redact init
# 0c. git filter-branch --msg-filter 'sed s/git-crypt/redact/g' \
#     --tree-filter path/to/_utilities/gitcrypt2redact.sh --prune-empty master

# 1. change gitattributes

find . -name .gitattributes | xargs sed -i '' -e 's/git-crypt/redact/g'

# 2. find new git-crypt collaborators, and switch them to redact collaborators

RED=../.redact
GA=.gitattributes
mkdir -p $RED .redact
find .git-crypt/keys/default/0 -name \*.gpg 2>/dev/null|
while read keyfn; do
    keyfile=${keyfn##*/}
    key=${keyfile%.gpg}
    key=$(echo $key | tr 'A-F' 'a-f')
    if [[ -f $RED/$key.key ]]; then
        cp $RED/$key.* .redact
        [[ -f $RED/$GA && ! -f .redact/$GA ]] && cp $RED/$GA .redact
    else
        echo $key
    fi
done | while read key; do
    redact access grant $key
    cp .redact/$key.* $RED
    [[ -f .redact/$GA && ! -f $RED/$GA ]] && cp .redact/$GA $RED
done
rm -rf .git-crypt

# 3. fix redact encryption
ENC=../.encoded
mkdir -p $ENC
git ls-files --stage | while read mode hash stage file; do
    EH=$ENC/$hash
    if [[ -f $EH ]]; then
        cp $EH $file
        # echo "already encrypted file: $file" >&2
        continue
    fi
    if ! git check-attr filter $file | grep -qF ': filter: unspecified'; then
        echo "new encrypted file: $file $hash" >&2
        if git-crypt smudge < "$file" > "$ENC/clean"; then
            redact git clean < "$ENC/clean" > "$EH"
            rm "$ENC/clean"
            cp $EH $file
        else
            echo "error decrypting $file:"
            cat $file
            echo "----"
            cat $$EH
        fi
    fi
done

git config --local --remove-section filter.redact 2>/dev/null || true
git config --local --remove-section filter.git-crypt 2>/dev/null || true
