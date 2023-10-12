## RIGO Account Metadata Schema

```json
{
  "name": "Account owner's name",
  "image": "URL pointing to an image related with the account. e.g) icon, logo...", 
  "image_data": "Raw SVG image data", 
  "description": "Introduction for an account", 
  "external_url": "Homepage, Github page etc.", 
  "attributes": [
    {
      "display_type": "string/number/date",
      "trait_type": "field name",
      "value": "field value"
    },
  ]
}
```

**References**

[ERC721](https://github.com/ethereum/EIPs/blob/master/EIPS/eip-721.md)

[Opensea metadata structure](https://docs.opensea.io/docs/metadata-standards#metadata-structure)