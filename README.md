# vk-fasting

A small command-line tracker for intermittent fasting, written in Go. Log your fasts, see the latest ones at a glance, and keep your data in a plain JSON file with an automatic backup to an external drive.

## Features

- Add, update, delete and undo fasts from the command line
- Prints a table of your last 10 fasts after every command
- Duration is calculated automatically from the start and end dates
- Ongoing fasts are supported: leave the end empty and it shows as `(ongoing)`
- Asks for confirmation before deleting anything
- Data is stored in a human-readable JSON file
- Safe saving: files are written atomically, so a crash can't leave a half-written file
- Best-effort backup to an external drive (a failed backup only prints a warning)

## Install

Requires Go 1.21 or newer.

```sh
git clone https://github.com/VkHyperNova/vk-fasting.git
cd vk-fasting
go build -o fasting .
```

## Usage

```
fasting [command] [id]
```

| Command            | What it does                                   |
| ------------------ | ---------------------------------------------- |
| *(none)*           | Show the last 10 fasts                         |
| `a`, `add`         | Add a new fast                                 |
| `u`, `update <id>` | Update the fast with the given ID              |
| `d`, `delete <id>` | Delete the fast with the given ID (asks first) |
| `undo`             | Delete the most recent fast (asks first)       |
| `help`, `-h`       | Show help                                      |

### Adding and updating

You are prompted for each field:

- **Start**: date in `DD.MM.YY` format, e.g. `28.10.26`. Defaults to today when adding.
- **End**: same format. Leave empty for a fast that is still going.
- **Weight**: free text, optional.

When updating, press Enter on any field to keep its current value. The duration is recalculated every time.

### Example output

```
Last 3 fasts

ID   START      END         DURATION   WEIGHT
--   -----      ---         --------   ------
1    24.10.26   26.10.26    48h00m     80
2    26.10.26   28.10.26    48h00m     79.5
3    28.10.26   (ongoing)              -
```

## Data and backup

Your fasts are saved to `DEVELOPMENT/FASTING/fasting.json`, relative to the folder you run the program from. The file looks like this:

```json
{
  "fastings": [
    {
      "id": 1,
      "start": "24.10.26",
      "end": "26.10.26",
      "duration": "48h00m",
      "weight": "80"
    }
  ]
}
```

After every save, the program tries to mount the drive labeled `VK DATA` (using [vkmount](https://github.com/VkHyperNova/vkmount)) and copies the file to `/media/veikko/VK DATA/DEVELOPMENT/FASTING/fasting.json`. If the drive isn't available, you get a warning and the local save still succeeds.

The paths, drive label and date format are constants at the top of `main.go` (`baseLocal`, `baseDB`, `driveLabel`, `driveMountPoint`, `timeLayout`), so change them there to fit your setup.

## Project layout

| File         | Contents                                                |
| ------------ | ------------------------------------------------------- |
| `main.go`    | Constants, argument handling and command dispatch       |
| `fastings.go`| Fast types, add/update/delete/undo and the input prompts |
| `storage.go` | Loading, saving, atomic file writes and the backup      |
| `report.go`  | The screen header and the table of recent fasts         |

## Tests

```sh
go test ./...
```

The tests use temporary directories and a stubbed drive mount, so they never touch your real data or the backup drive.