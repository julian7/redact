# To be implemented

- key exchange
  - allow keybase's encryption to be used
- fix encryption issues
  - encrypted what shouldn't been
  - not encrypted which should have been
  - re-encrypt with a newer key (use other option than --fix)
- generating new key should rekey encrypted files
- git smudge should alert if unlock is needed (when newer key is needed than the latest available)
