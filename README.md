# e12pack - ArmA Addon Packer

e12pack automates the process of packaging multiple folders to pbos and moving them to the mod/addons folder to get loaded by ArmA.

## Usage

```bash
$ e12pack.exe -pack="C:\Users\blang\mymod"
```

Or better install the shell extension by executing `installShellExt.bat` as Administrator.
After that, right-click on the `mymod` folder and choose `E12Pack` to execute the command above.


The directory structure looks like this:

```
mymod 
 |
 |- mymod_main/
 |   |- config.cpp
 |   |- ...
 |- mymod_events/
 |   |- config.cpp
 |   |- stringtable.xml
 |   |- ...
 |- .e12pack
 |- .e12pack_settings
```

## Files

The `.e12pack` file defines which folders get packed:
```
mymod_main
mymod_events
```


The `.e12pack_settings` file holds user defined settings:
```
output = "C:\\Program Files (x86)\\Steam\\SteamApps\\common\\Arma 3\\@mymod\\Addons"

[packer]
name = "PBOManager"
path = "C:\\Program Files\\PBO Manager v.1.4 beta\\PBOConsole.exe"
```

The `.e12pack_settings` file should be included in your `.gitignore` file.

In the above sample, the command would result in `mymod_main.pbo` and `mymod_events.pbo` created in `@mymod\\Addons\`

### Packer Section
Currently only `PBOManager` is supported, with further packers in mind.

## Contribution

Feel free to make a pull request. For bigger changes create a issue first to discuss about it.


## License

See [LICENSE](LICENSE) file.