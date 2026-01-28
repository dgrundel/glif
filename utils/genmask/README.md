# genmask

Generate a `.mask` file from a `.sprite` file by mapping non-space characters to a fill rune and spaces to a transparent rune.

## Usage

```
go run ./utils/genmask path/to/sprite.sprite
```

## Options

```
--out path/to/output.mask
--fill w
--transparent .
```

## Example

```
go run ./utils/genmask demos/duck/assets/whale.sprite
```

This writes `demos/duck/assets/whale.mask` by default.
