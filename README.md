Ekklesion Tenancy
=================

A Caddy module to extract dynamic information of tenants at runtime. Suited for multi-tenant PHP applications.

## Why this

Traditionally, there has been two ways of achieving multi-tenancy in PHP applications. I call these the *runtime*
approach and the *process* approach.

In the *runtime* approach, the PHP script determines at-runtime the tenant in some http middleware layer that is then passed 
down to the application. This makes the tenant pollute all of the application logic, and brings tremendous complexity,
particularly for storage. This is because you either have to (a) pollute every database table with a `tenant_id` column or
(b) connect to databases at runtime by pulling secrets and other information from a central tenant repository. While the
first approach is easier to manage, it's harder to scale. The second approach scales better, but it's tremendously
insecure (because the PHP process has access to all the tenant information) and more complex to implement (because before
being able to connect to any external service, tenant information must be retrieved). One of the benefits though, is that
you can serve multiple tenants with a single instance.

In the *process* approach, you basically deploy one process per tenant. This means that you will have one container (or
some other workload) running per tenant. This workload will have the traditional environment variables set to the
corresponding tenant preferences. This has the advantage of being extremely convenient, because most PHP applications and
frameworks are made to work in this scenario and the application is somewhat tenant unaware (the environment variables
really determine the tenant). This approach also physically separates storage from one tenant to the other with zero
effort. The problem is that is wasteful, specially if you are in the business of developing back-office panels that
don't really have much traffic. You'll be allocating resource and memory that is likely not going to be used much, 
requiring you to spend more in compute costs for each tenant you have.

This Caddy module allows you to have the best of both approaches, with very little drawbacks. It makes it possible
to leave your PHP backend applications unchanged (along with their environment variables) but dynamically inject parameters
at runtime based on information obtained from an external source. The external source could be as complex as a Vault server
or as simple as json files in a folder (both approaches are implemented here for reference). Then, those parameters can
be dynamically replaced in your caddy configuration and be injected as fastcgi environment variables.

This is an example of how your caddy file would look like:

```caddyfile
*.cloud.myapp.dev {
    @site {
        # This basically means the site information is in json files in this directory
        # The json file must be named as the host + the ".json" extension. Ex: one.cloud.myapp.dev.json
        # If the matcher finds such site, the site information gets injected.
        site_info json+dir:/myapp/sites?watch=true
    }
    
    # You can create further matchers based on the injected variables
    # `site_info.status` and `site_info.id` are stable variables provided by this matcher
    @enabled {
        vars {site_info.status} enabled
    }

    @disabled {
        vars {site_info.status} disabled
    }

    encode gzip

    handle @site {
        php_fastcgi @enabled fpm:9000 {
            root /myapp/public
            # Everything under params becomes a caddy variable that you can use later
            env APP_SECRET {myapp.env.app_secret}
            env DATABASE_URL {myapp.env.db_url}
            env MAILER_DSN {myapp.env.mailer}
        }

        error @disabled "Your site is disabled" 500
    }

    error "Not Found" 404
}
```

## Development

Get started by running `docker compose up -d`.

Then, go to [http://one.cloud.localhost/index.php](http://one.cloud.localhost/index.php)