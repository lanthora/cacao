# Cacao

Candy Server with WebUI

## Build

```bash
# build a binary that runs natively
make

# build multiple platform binaries
make all
```

## Run

```bash
# loglevel=[info] listen=[:80] storage=[.]
cacao

# loglevel=[debug] listen=[127.0.0.1:8080] storage=[/var/lib/cacao]
cacao --loglevel=debug --listen=127.0.0.1:8080 --storage=/var/lib/cacao
```

## Geographic database updates

Administrators can configure separate HTTP(S) download URLs and update intervals
for IPDB and MMDB in Settings. URLs must return raw database files (not archives).
An empty URL or an interval of 0 disables automatic updates. Save settings before
using **Update now**. Automatic updates are checked at startup and hourly; failures
are retried on the next check. Downloads are limited to 256 MiB and five minutes.

HTTPS downloads use the system's trusted CA certificates. The Docker image includes
a CA bundle; native installations need the operating system's CA certificates.
For a private CA or an HTTPS inspection proxy, mount a PEM bundle containing the
required trusted roots and set `SSL_CERT_FILE` to its path in the container.

Validated versions are published under `storage/geodb/ipdb/` and
`storage/geodb/mmdb/`. Failed downloads leave the current version usable. Startup
recovers the newest valid published version; old managed versions are cleaned up.
Existing `.ipdb` / `.mmdb` files in the storage root remain as fallbacks and are
never deleted by the updater. Do not share one storage directory between multiple
running Cacao processes.

Location lookup priority remains IPDB, then MMDB, then ipinfo. IPDB files must
support the CN language and contain `country_code` plus `region_name` or
`city_name`. MMDB files must support GeoIP2 City lookups. Updates affect subsequent
lookups; existing device locations are refreshed by normal device activity.
