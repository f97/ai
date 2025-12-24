# Using API to Control & Extend One API
> Welcome to submit PR to add your extension projects here.

For example, although One API does not directly support payment, you can implement payment functionality through the system's extension API.

Or if you want to customize channel management strategies, you can also use the API to disable and enable channels.

## Authentication
One API supports two authentication methods: Cookie and Token. For Token, refer to the figure below to obtain it:

![api-token](https://user-images.githubusercontent.com/39998050/233837884-64a1cdb5-0c3e-4347-84e1-bf56bb8f3e3d.png)

Subsequently, add `Authorization: Bearer YOUR_TOKEN` to the request header.

## Usage Examples

### Create Channel
```bash
curl --location --request POST 'http://localhost:3000/api/channel/' \
--header 'Authorization: Bearer YOUR_TOKEN' \
--header 'Content-Type: application/json' \
--data-raw '{
  "type": 1,
  "name": "OpenAI Channel",
  "key": "sk-xxx",
  "base_url": "https://api.openai.com"
}'
```

### Update Channel Status
```bash
curl --location --request PUT 'http://localhost:3000/api/channel/' \
--header 'Authorization: Bearer YOUR_TOKEN' \
--header 'Content-Type: application/json' \
--data-raw '{
  "id": 1,
  "status": 2
}'
```

Replace `YOUR_TOKEN` with your actual API token.

## API Reference
For complete API reference, see the code or use browser developer tools to inspect network requests.
