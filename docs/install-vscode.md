# Install Bitrise MCP Server in VS Code

## Prerequisites
- [VS Code](https://code.visualstudio.com/Download) installed
- [Create a Bitrise API Token](https://devcenter.bitrise.io/api/authentication):
   - Go to your [Bitrise Account Settings/Security](https://app.bitrise.io/me/account/security).
   - Navigate to the "Personal access tokens" section.
   - Copy the generated token.
- For local setup: [Go](https://go.dev/) (>=1.23) installed

## Remote Server Setup (Streamable HTTP)

Follow [VS Code | Add an MCP server](https://code.visualstudio.com/docs/copilot/customization/mcp-servers#_add-an-mcp-server) and add the following configuration to your settings:

```json
{
	"servers": {
		"bitrise": {
			"type": "http",
			"url": "https://mcp.bitrise.io",
			"headers": {
			  "Authorization": "Bearer ${input:bitrise-token}"
			}
		}
	},
	"inputs": [
		{
			"id": "bitrise-token",
			"type": "promptString",
			"description": "Bitrise token",
			"password": true
		}
	]
}
```

Save the configuration. VS Code will automatically recognize the change and load the tools into Copilot Chat.

## Local Server Setup (Go required)

Do the same as above, but use the following configuration instead:

```json
{
	"servers": {
		"bitrise-local": {
			"type": "stdio",
			"command": "go",
			"args": [
				"run",
				"github.com/bitrise-io/bitrise-mcp@v2"
			],
			"env": {
				"BITRISE_TOKEN": "${input:bitrise-token}"
			}
		}
	},
	"inputs": [
		{
			"id": "bitrise-token",
			"type": "promptString",
			"description": "Bitrise token",
			"password": true
		}
	]
}
```

## Advanced configuration

See [Tools](tools.md) for enabling/disabling specific API groups.
