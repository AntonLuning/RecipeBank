## Endpoints

If not stated otherwise, all data (request and response body) will be a JSON object.

### GET /recipe

Fetches a list (JSON array) with all existing recipes. Each recipe data include `title`, `image`, and `type`.

> `type` is an array of one or more types the recipe is categorized into.

**Query params:**
- *TODO: filter options*

### GET /recipe/{UUID}

Fetches a single recipe based on its `UUID` with detailed (all available) data.

### POST /recipe

Create a new recipe. All recipe data will be optional with default values if omitted.

### PUT /recipe/{UUID}

Update an existing recipe based on its `UUID`. Request can include one or more fields to be updated.
