# Redact stands on the shoulders of giants

Redact project doesn't try to reinvent the wheel. It is based on great ideas put together by other projects, and it uses the already great wheels developed by the Go community and puts them together in the best way possible. Without these giants, this project would not be possible. Please make sure to check them out and thank them for all of their hard work.

Thank you for the following **GIANTS**:

* [git-crypt](https://github.com/AGWA/git-crypt): this project started as a go implementation of git-crypt, and most important commands try to mimic git-crypt behavior.
* [vault](https://github.com/hashicorp/vault): I've turned to vault for inspiration, and implementation advice for the replacement symmetric encryption.
* [git](https://github.com/git/git): this project wouldn't see the light if there was no git.
* [GnuPG](https://gnupg.org/): the project took the torch over from PGP, which democratized secret communication for civilians. By now the majority of its features are obsoleted by newer technologies, it is still a strong bastion of civil privacy.

This project uses go modules, so please have a look of [go.sum](go.sum) for all the Go **GIANTS** this project is stand upon.
