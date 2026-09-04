---
setup:
  dotenv: [setup_dotenv.env]
  envVars:
    - name: FROM_SETUP_VARS
      value: setup-vars
env:
  FROM_DOTENV: from-fixture
---

# setup: environment

`setup:` loads dotenv files and explicit `envVars` into every fixture in this
document. Assertions use `printenv` rather than `echo $VAR`, because fixture args
are templated before the shell runs — a `$VAR` in an argument would assert the
template variable rather than the child process's actual environment.

## Values reach the child process

| Name             | CLI      | Args            | CEL Validation                 |
|------------------|----------|-----------------|--------------------------------|
| dotenv only      | printenv | ONLY_IN_DOTENV  | stdout.trim() == "dotenv"      |
| explicit env var | printenv | FROM_SETUP_VARS | stdout.trim() == "setup-vars"  |

## Precedence: fixture env beats setup env

`FROM_DOTENV` is set to `from-dotenv` by the dotenv file and to `from-fixture` by
the document's own `env:`. The fixture wins.

| Name             | CLI      | Args        | CEL Validation                  |
|------------------|----------|-------------|---------------------------------|
| fixture env wins | printenv | FROM_DOTENV | stdout.trim() == "from-fixture" |

## The setup directory is exported

With no `checkout:` the prepared directory is the markdown file's own directory,
so `$SETUP_DIR` and the working directory agree.

### command: SETUP_DIR is the working directory

```bash
test "$(printenv SETUP_DIR)" = "$(pwd)"
```
