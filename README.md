Tile Configuration Convertor
---

## Motivation
To automate deployment of any of the product tiles shipped by Pivotal, the configuration parameters are the ones that are hard to fetch. This tool was written to ease the fetching of the properties of the product tile, after the tile has been uploaded and staged.

## How this works
Once the product is uploaded and staged, the platform engineer can use the [om cli](https://github.com/pivotal-cf/om/releases).

To download all the available properties for a given tile execute:

```
om -t OPS-MANAGER-URL -u USERNAME -p PASSWORD curl -p /api/v0/staged/products/$PRODUCT_GUID/properties > properties.json
```

To download all the available resources for a given tile execute:

```
om -t OPS-MANAGER-URL -u USERNAME -p PASSWORD curl -p /api/v0/staged/products/$PRODUCT_GUID/resources > resources.json
```

To download all the available errands for a given tile execute:

```
om -t OPS-MANAGER-URL -u USERNAME -p PASSWORD curl -p /api/v0/staged/products/$PRODUCT_GUID/errands > errands.json
```

These commands produce a **json** file.

In-order to get the properties that can be configured, execute:

```
tile-config-convertor -c properties -i properties.json -o properties.yml -ov properties-vars.yml
```

In-order to get the resources that can be configured, execute:

```
tile-config-convertor -c resources -i resources.json -o resources.yml -ov resources-vars.yml
```

In-order to get the errands that can be configured, execute:

```
tile-config-convertor -c errands -i errands.json -o errands.yml -ov errands-vars.yml
```

In-order to get the networks and az template that could be configured, execute:

```
tile-config-convertor -c network-azs -o network-azs.yml -ov network-azs-vars.yml
```

You can now paste the output contents into the params.yml for the given tile and fly them using the [install-product pipeline](https://github.com/rahulkj/pcf-concourse-pipelines/tree/master/pipelines/install-product)

**NOTE: New features will be added as and when they are identified**
