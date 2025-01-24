# go-radvd-manager

<summary>
    <code>[GET|DELETE]</code> 
    <code><b>/rest/data/radvd:instances</b></code></br>
    <code>[GET|PUT|POST|DELETE]</code> 
    <code><b>/rest/data/radvd:instances/{instance}</b></code>
</summary>

## Endpoints

- `[GET]/rest/data/radvd:instances`: Get all instances information.
  ```
  $ curl -s http://localhost:12345/rest/data/radvd:instances | jq .
  ```
  ```json
  [
    {
      "id": 0, // Reserved for default process
      "pid": 0, // Zero means unknown
      "router_id": "",
      "name": "docker0",
      "adv_send_advert": true,
      "min_rtr_adv_interval": 0,
      "max_rtr_adv_interval": 0,
      "adv_managed_flag": false,
      "adv_other_config_flag": false,
      "adv_default_lifetime": 0,
      "adv_default_preference": "",
      "prefixes": null,
      "rdnss": null,
      "routes": [
        {
          "route": "2001:db8:1::1/128",
          "adv_route_lifetime": 300,
          "adv_route_preference": "medium"
        },
        {
          "route": "2001:db8:2::/64",
          "adv_route_lifetime": 300,
          "adv_route_preference": "medium"
        }
      ],
      "clients": null
    }
  ]
  ```

- `[POST|PUT]/rest/data/radvd:instances/{instance}`: Start/Update radvd instance with specified id.
  ```
  $ curl -X POST -H "Content-Type: application/yang-data+json" -d @testdata/instance.json http://localhost:12345/restconf/data/radvd:interfaces/5
  $ curl -s http://localhost:12345/rest/data/radvd:instances/5 | jq 
  ```
  > Note: The values of `{instance}` and `id:`in testdata must be the same.
  ```json
  {
    "id": 5,
    "pid": 184779,
    "router_id": "::1",
    "name": "docker0",
    "adv_send_advert": true,
    "min_rtr_adv_interval": 3,
    "max_rtr_adv_interval": 10,
    "adv_managed_flag": false,
    "adv_other_config_flag": false,
    "adv_default_lifetime": 0,
    "adv_default_preference": "medium",
    "prefixes": [
      {
        "prefix": "2001:db8::/64",
        "adv_on_link": true,
        "adv_autonomous": true,
        "adv_router_addr": true,
        "adv_valid_lifetime": 86400
      },
      {
        "prefix": "2001:db8:abcd::/64",
        "adv_on_link": true,
        "adv_autonomous": false,
        "adv_router_addr": false,
        "adv_valid_lifetime": 43200
      }
    ],
    "rdnss": [
      {
        "address": "2001:db8::1",
        "adv_rdnss_lifetime": 1800
      },
      {
        "address": "2001:db8::2",
        "adv_rdnss_lifetime": 1500
      }
    ],
    "routes": [
      {
        "route": "2001:db8:abcd::/48",
        "adv_route_lifetime": 300,
        "adv_route_preference": "medium"
      },
      {
        "route": "2001:db8::/32",
        "adv_route_lifetime": 600,
        "adv_route_preference": "high"
      }
    ],
    "clients": [
      "fe80::1"
    ]
  }
  ```

- `[DELETE]/rest/data/radvd:instances`
  - Delete all radvd instances.
    ```
    $ curl -X DELETE http://localhost:12345/restconf/data/radvd:interfaces
    ```
- `[DELETE]/rest/data/radvd:instances/{intstance}`
  - Delete specified radvd instance.
    ```
    $ curl -X DELETE http://localhost:12345/restconf/data/radvd:interfaces/5
    ```

## Responses
> | http method  |  request body  | response body |
> |--------------|----------------|---------------|
> | `POST`       |  *JSON data*   | - none -      |
> | `PUT`       |  *JSON data*   | - none -      |
> | `GET`       |  - none -      | *JSON data*   |
> | `DELETE`     |  - none -      | - none -      |


<!-- ########################################################### -->
## HTTP response codes
> | http code |  reason for code    |
> |-----------|---------------------|
> | 200       | success             |
> | 400       | invalid request     |
> | 404       | data does not exist |
> | 500       | internal error      |