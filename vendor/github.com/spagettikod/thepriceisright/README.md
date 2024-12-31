<p><a href="https://www.elprisetjustnu.se"><img src="https://ik.imagekit.io/ajdfkwyt/hva-koster-strommen/elpriser-tillhandahalls-av-elprisetjustnu_ttNExOIU_.png" alt="Elpriser tillhandahålls av Elpriset just nu.se" width="200" height="45"></a></p> 

# The Price Is Right

The Price Is Right is a command line tool to check if the current price of electricity is equal to or below a given price. The tool can be used when scripting things that depend on the price of electricity.

Use cases:
* triggering the charge (or discharge) of residential batteries when the price reaches a certain point
* shutting down / powering up devices that consume lots of electricity

This tool only works for the Swedish electricity market and uses the API from [Elpriset just nu.se](https://www.elprisetjustnu.se).

## Download
### Linux
```
curl -Ls https://github.com/spagettikod/ThePriceIsRight/releases/download/v1.1.0/tpir1.1.0.linux-amd64.tar.gz | tar xz
```

### MacOS
```bash
curl -Ls https://github.com/spagettikod/ThePriceIsRight/releases/download/v1.1.0/tpir1.1.0.macos-arm64.tar.gz | tar xz
```

## Usage
The Price Is Right require two parameters.
* electricity area code that you want to check the price for. Valid values are `SE1`, `SE2`, `SE3` or `SE4`.
* maximum acceptable price in Swedish krona per kWh without taxes and other charges.

```sh
tpir SE3 2.35
```

If the price is below or equals `2.35` SEK/kWh when we run the above example `tpir` exits with the exit code `0`. If the price would happen to be above `2.35` SEK/kWh it exits with `1`. If an error would occur it exits with `2`.

Calling `tpir` will load todays prices from the cache. If there are no cache files for the given electricity area code or if the cached price list has expired the cache is updated by calling the REST API.

### Example
Here is a more complete example which can be used as a template for a script that will check the price of electricity every 5 minutes and run different commands depending on the price in area `SE3` are 0.87 SEK/kWH or below. 
```sh
#!/bin/sh

while true; do
    if $(tpir SE3 0.87); then
        echo "Price looks fine!"
    else
        echo "Electricity is way too expensive!"
    fi
    sleep 300
done
```

### Caching
Daily prices fetched from the REST API are cached locally until they expire for the day. When they expire they are refreshed by calling the REST API again. There is a file for each electricity area code.

On Linux the cache files are stored at `$XDG_CACHE_HOMEthepriceisright/XX_cache.json` or `$HOME/.cache/thepriceisright/XX_cache.json`.

On MacOS the cache files are stored at `$HOME/Library/Caches/thepriceisright/XX_cache.json`.
