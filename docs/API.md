# go-radvd-manager

<summary>
    <code>[GET|PUT|POST|DELETE]</code> 
    <code><b>restconf/data/radvd:interfaces/{instance}</b></code>
</summary>

## Endpoints

- GET: Get all interfaces information.
  - `restconf/data/radvd:interfaces/`
  ```
  $ curl -s http://localhost:8888/restconf/data/radvd:interfaces | jq .
  ```
  ```json
  [
    {
      "instance": 0,
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

- GET: Get specified interface information.
  - `restconf/data/radvd:interfaces/{instance}`
  ```
  $ curl -s http://localhost:8888/restconf/data/radvd:interfaces/2 | jq .
  ```

  ```json
  {
    "instance": 2,
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

- POST / PUT: Start/Update radvd instance with specified id.
  - `restconf/data/radvd:interfaces/{instance}`
  ```
  $ curl -X POST -H "Content-Type: application/yang-data+json" -d @sample2.json http://localhost:8888/restconf/data/radvd:interfaces/2
  ```
  > Note: The values of `{instance}` and `instance:` must be the same.
  ```json
  ### sample.json ###
  {
    "instance": 2,
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

- DELETE: 
  - Delete all radvd instances.
    - `restconf/data/radvd:interfaces`
    ```
    $ curl -X DELETE http://localhost:8888/restconf/data/radvd:interfaces
    ```
  - Delete specified radvd instance.
    - `restconf/data/radvd:interfaces/{intstance}`
    ```
    $ curl -X DELETE http://localhost:8888/restconf/data/radvd:interfaces/2
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