## cozy-stack serve

Starts the stack and listens for HTTP calls

### Synopsis

Starts the stack and listens for HTTP calls
It will accept HTTP requests on localhost:8080 by default.
Use the --port and --host flags to change the listening option.

The SIGINT signal will trigger a graceful stop of cozy-stack: it will wait that
current HTTP requests and jobs are finished (in a limit of 2 minutes) before
exiting.

If you are the developer of a client-side app, you can use --appdir
to mount a directory as the application with the 'app' slug.


```
cozy-stack serve [flags]
```

### Examples

```
The most often, this command is used in its simple form:

	$ cozy-stack serve

But if you want to develop two apps in local (to test their interactions for
example), you can use the --appdir flag like this:

	$ cozy-stack serve --appdir appone:/path/to/app_one,apptwo:/path/to/app_two

```

### Options

```
      --allow-root                                 Allow to start as root (disabled by default)
      --appdir strings                             Mount a directory as the 'app' application
      --assets string                              path to the directory with the assets (use the packed assets by default)
      --couchdb-url string                         CouchDB URL (default "http://localhost:5984/")
      --csp-allowlist string                       Add domains for the default allowed origins of the Content Secury Policy
      --dev                                        Allow to run in dev mode for a prod release (disabled by default)
      --disable-csp                                Disable the Content Security Policy (only available for development)
      --doctypes string                            path to the directory with the doctypes (for developing/testing a remote doctype)
      --downloads-url string                       URL for the download secret storage, redis or in-memory
      --flagship-apk-certificate-digests strings   SHA-256 hash (base64 encoded) of the flagship app's signing certificate on android (default [u2eUUnfB4Y7k7eqQL7u2jiYDJeVBwZoSV3PZSs8pttc=])
      --flagship-apk-package-names strings         Package name for the flagship app on android (default [io.cozy.drive.mobile,io.cozy.flagship.mobile])
      --flagship-apple-app-ids strings             App ID of the flagship app on iOS (default [3AKXFMV43J.io.cozy.drive.mobile,3AKXFMV43J.io.cozy.flagship.mobile])
      --fs-default-layout int                      Default layout for Swift (2 for layout v3) (default -1)
      --fs-url string                              filesystem url (default "file:///home/runner/work/cozy-stack/cozy-stack/storage")
      --geodb string                               define the location of the database for IP -> City lookups (default ".")
  -h, --help                                       help for serve
      --jobs-url string                            URL for the jobs system synchronization, redis or in-memory
      --konnectors-cmd string                      konnectors command to be executed
      --konnectors-oauthstate string               URL for the storage of OAuth state for konnectors, redis or in-memory
      --lock-url string                            URL for the locks, redis or in-memory
      --log-level string                           define the log level (default "info")
      --log-syslog                                 use the local syslog for logging
      --mail-alert-address string                  mail address used for alerts (instance deletion failure for example)
      --mail-disable-tls                           disable starttls on smtp (default true)
      --mail-host string                           mail smtp host (default "localhost")
      --mail-local-name string                     hostname sent to the smtp server with the HELO command (default "localhost")
      --mail-noreply-address string                mail address used for sending mail as a noreply (forgot passwords for example)
      --mail-noreply-name string                   mail name used for sending mail as a noreply (forgot passwords for example) (default "My Cozy")
      --mail-password string                       mail smtp password
      --mail-port int                              mail smtp port (default 25)
      --mail-reply-to string                       mail address used to the reply-to (support for example)
      --mail-use-ssl                               use ssl for mail sending (smtps)
      --mail-username string                       mail smtp username
      --mailhog                                    Alias of --mail-disable-tls --mail-port 1025, useful for MailHog
      --move-url string                            URL for the move wizard (default "https://move.cozycloud.cc/")
      --onlyoffice-inbox-secret string             Secret used for signing requests to the OnlyOffice server
      --onlyoffice-outbox-secret string            Secret used for verifying requests from the OnlyOffice server
      --onlyoffice-url string                      URL for the OnlyOffice server
      --password-reset-interval string             minimal duration between two password reset (default "15m")
      --rate-limiting-url string                   URL for rate-limiting counters, redis or in-memory
      --realtime-url string                        URL for realtime in the browser via webocket, redis or in-memory
      --remote-allow-custom-port                   Allow to specify a port in request files for remote doctypes
      --sessions-url string                        URL for the sessions storage, redis or in-memory
      --subdomains string                          how to structure the subdomains for apps (can be nested or flat) (default "nested")
      --vault-decryptor-key string                 the path to the key used to decrypt credentials
      --vault-encryptor-key string                 the path to the key used to encrypt credentials
```

### Options inherited from parent commands

```
      --admin-host string   administration server host (default "localhost")
      --admin-port int      administration server port (default 6060)
  -c, --config string       configuration file (default "$HOME/.cozy.yaml")
      --host string         server host (default "localhost")
  -p, --port int            server port (default 8080)
```

### SEE ALSO

* [cozy-stack](cozy-stack.md)	 - cozy-stack is the main command

