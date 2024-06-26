---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "rockset_gcs_collection Resource - rockset"
subcategory: ""
description: |-
  Manages a collection with an GCS source attached.
---

# rockset_gcs_collection (Resource)

Manages a collection with an GCS source attached.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Unique identifier for the collection. Can contain alphanumeric or dash characters.
- `workspace` (String) The name of the workspace.

### Optional

- `description` (String) Text describing the collection.
- `ingest_transformation` (String) Ingest transformation SQL query. Turns the collection into insert_only mode.

When inserting data into Rockset, you can transform the data by providing a single SQL query, 
that contains all of the desired data transformations. 
This is referred to as the collection’s ingest transformation or, historically, its field mapping query.

For more information see https://rockset.com/docs/ingest-transformation/
- `retention_secs` (Number) Number of seconds after which data is purged. Based on event time.
- `source` (Block Set) Defines a source for this collection. (see [below for nested schema](#nestedblock--source))
- `storage_compression_type` (String) RocksDB storage compression type. Possible values: ZSTD, LZ4.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `wait_for_collection` (Boolean) Wait until the collection is ready.
- `wait_for_documents` (Number) Wait until the collection has documents. The default is to wait for 0 documents, which means it doesn't wait.

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--source"></a>
### Nested Schema for `source`

Required:

- `bucket` (String) GCS bucket containing the target data.
- `format` (String) Format of the data. One of: json, csv, xml. xml and csv blocks can only be set for their respective formats.
- `integration_name` (String) The name of the Rockset GCS integration.

Optional:

- `csv` (Block Set, Max: 1) (see [below for nested schema](#nestedblock--source--csv))
- `prefix` (String) Simple path prefix to GCS key.
- `xml` (Block Set, Max: 1) (see [below for nested schema](#nestedblock--source--xml))

<a id="nestedblock--source--csv"></a>
### Nested Schema for `source.csv`

Optional:

- `column_names` (List of String) The names of the columns.
- `column_types` (List of String) The types of the columns.
- `encoding` (String) Can be one of: UTF-8, ISO_8859_1, UTF-16.
- `escape_char` (String) Escape character removes any special meaning from the character that follows it. Defaults to backslash.
- `first_line_as_column_names` (Boolean) If the first line in every object specifies the column names.
- `quote_char` (String) Character within which a cell value is enclosed. Defaults to double quote.
- `separator` (String) A single character that is the column separator.


<a id="nestedblock--source--xml"></a>
### Nested Schema for `source.xml`

Optional:

- `attribute_prefix` (String) Tag to differentiate between attributes and elements.
- `doc_tag` (String) Tags with which documents are identified
- `encoding` (String) Encoding in which data source is encoded.
- `root_tag` (String) Tag until which xml is ignored.
- `value_tag` (String) Tag used for the value when there are attributes in the element having no child.



<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
