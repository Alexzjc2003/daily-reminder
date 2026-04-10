# Daily-reminder

## Introduction

Daily-reminder, or dr, is a reminder designed to work with Linux login message or .bashrc to log remind message on login or bash start.

## Usage

### Init

For the first time using dr, run

```bash
dr init [-d <directory>]
```

in shell, which will create a `.daily-reminder/` folder under your home directory by default. If the directory already exits, this will exit with an error.
`-d` flag will make dr create the folder under specified directory. I.e., `dr init -d $HOME` is equivalent to `dr init`.

If `-d` flag is specified, you will need to set a DR_DIR in your environment variables for dr to find the directory in future command executions.

```bash
# i.e., if you run
dr init -d /directory/to/init/dr

# then, you may set in ~/.bashrc
export DR_DIR="/directory/to/init/dr"
```

> From now on, we will use `DR_DIR` to indicate this directory, and `DR_PATH` to indicate `DR_DIR/.daily-reminder`

### Remember a date

```bash
dr remember <name> <date>
```

To remember a date, you need to have a name for that date. And for the date part, dr only accept the format of YYYY/MM/DD (i.e., `time.Parse("2006/01/02", date)`).

> If you are using `tm`, this would be `tm -f "%Y/%m/%d"`
> We accept some fancy date, e.g., 0000/02/18, to indicate that happens each year.

### Query a date

There are mainly 2 query modes in dr, _regular_ mode and _expand_ mode. By default, `dr query` queries in _regular_ mode. You can specify _expand_ mode by calling `dr query -x`.

#### _regular_ query and _expand_ query

As mentioned above, dr accepts some fancy date pattern, such as `0000/02/18` to indicate something happens each year on Feb 18th. If we are querying a date range, e.g., `dr query --from="2026/01/01" --to="2026/12/31"`, `0000/02/18` will not be included. While in _expand_ mode, i.e., `dr query -x --from="2026/01/01" --to="2026/12/31"`, `0000/02/18` shall be included, as `2026/02/18` fulfills the query.

### Forget a date

```bash
dr forget <name>
```

### Check memory

```bash
dr memory
```

### Remind me

```bash
dr remind
```

## Dates

Under `DR_PATH` a lies a `dates` file, which remembers all the dr-dates. A dr-date is a date with a name. The format should be like

```
DATES := DATES [DR-DATE] EOL(\n)
DR-DATE := DATE : NAME(string) ; TRAITS
TRAITS := TRAIT(string), TRAITS | TRAIT(string)
DATE := YYYY/MM/DD(string)
```
