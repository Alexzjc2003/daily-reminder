# daily-reminder

## lang

`lang` module is a language for dr. It provides a way to use dr as a language and run `*.dr` files (or, `dr-script`s).

## Usage

To run a `.dr` file, e.g., `test.dr`, call 

```sh
dr run test.dr
```

> Actually we don't check for suffixes, which means you may try to run any file with `dr run` 

## Spec

### dr-lang cmd

Typically, a `dr-script` should be several `dr-cmd`s.

```
DR_SCRIPT := []DR_CMD
```

A `dr-cmd` is a cmd with its args, ended with a `;`.

```
DR_CMD := CMD []ARGS EOL(\n)
```

### variable

Variables come in 2 ways, manually written in scripts, or produced by [expression](###expression).

There are several kinds of variable. Expression produced values have fixed kinds, according to the expression. But for written ones, they can be decided as follows.

> Note that the following prefix rules only apply to dr-script. If args are in dr-cmds, they should follow cmd rules, to make more sense.

#### kinds

##### String

Anything not specifically marked is treated as `String`.

##### Number 

`Number`s are `int64`s. No prefix is needed for number parsing. E.g., `set $urmom 69`.

> A segment is first tried to be parsed as a number. If a segment is able to be parsed as a number, it will always be parsed as a number.

##### Date

`Date` is basically `ReminderDate`, and starts with "@" in dr-script. The format follows the same rule introduced in [INTRO](../INTRO.md) 

> If all parsing except string failed(without any prefix, prefixes force a specific parsing), a string parsing will always be a fallback.

##### Object

`Object` is basically a `map[string]Variable`. To fetch its fields, use dot(`.`), like `$object.value`. 

#### setting and getting a variable

```
set $recent query -x --from="today" --to="3 days later";
print $recent;
```

#### variable table

A variable table is a map that dr-lang holds at runtime. It is defined as follow:

```go
type VariableTable map[string]Variable

var vt VariableTable
```



### expression

## Cmd

Note that dr-lang does not cover all the dr-cli commands, but for those of the same name, they share similar usage. 

For now, supported dr-lang cmds are:
- remember
- query

> Note that args for cmds follow their own rules.