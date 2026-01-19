# Install Bitrise MCP Server in AWS Kiro

## Prerequisites
- AWS Kiro IDE installed
- [Create a Bitrise API Token](https://devcenter.bitrise.io/api/authentication):
  - Go to your [Bitrise Account Settings/Security](https://app.bitrise.io/me/account/security)
  - Navigate to the "Personal access tokens" section
  - Copy the generated token

## Installation via Kiro Power

AWS Kiro supports installing the Bitrise MCP server as a Power, which provides automatic activation based on context and keywords.

### Steps

1. **Open the Powers Panel**
   - In Kiro IDE, open the Powers panel from the sidebar

2. **Add Power from GitHub**
   - Click on "Add power from GitHub"

3. **Enter the Repository URL**
   - Enter the following URL:
   ```
   https://github.com/bitrise-io/bitrise-mcp/tree/main/power
   ```

4. **Set Up Authentication**
   - Before starting Kiro, set the `BITRISE_TOKEN` environment variable in your shell profile (`.zshrc` or `.bashrc`):
     ```bash
     export BITRISE_TOKEN="your-actual-token-here"
     ```
   - Restart your terminal or run `source ~/.zshrc` (or `source ~/.bashrc`)
   - Start Kiro - it will read the environment variable on startup
   - When Kiro starts, a popup may ask if you trust the environment variable - accept it to allow access

5. **Verify Installation**
   - The Bitrise Power should now appear in your Powers list
   - It will automatically activate when you mention keywords like "bitrise", "build", "ci", "cd", "mobile", "ios", "android", etc.

## Usage

Once installed, the Bitrise Power will automatically activate when relevant. You can:

- Manage Bitrise apps
- Trigger and monitor builds
- Handle build artifacts
- Manage workspaces and teams
- Configure pipelines
- Set up release management

The power provides access to all 63 Bitrise tools. For a complete list of available tools and their parameters, refer to the [tools documentation](/docs/tools.md).

## Advanced Configuration

You can limit the tools exposed by configuring API groups. This is useful for optimizing token usage or focusing on specific functionality.

Available API groups:
- `apps` - App management
- `builds` - Build operations
- `artifacts` - Artifact management
- `workspaces` - Workspace management
- `pipelines` - Pipeline operations
- `outgoing-webhooks` - Webhook configuration
- `cache-items` - Cache management
- `release-management` - Release and distribution
- `group-roles` - Role management
- `account` - User account operations
- `read-only` - Read-only operations

By default, all groups are enabled. To customize, modify the Power configuration after installation.

## Troubleshooting

### Environment Variable Not Working

Kiro IDE can only read environment variables that are set in your shell profile (`.zshrc` or `.bashrc`) **before** Kiro starts. Unlike VS Code, Kiro does not prompt you to enter the token value - it expects the environment variable to already be available.

**Option 1: Set in Shell Profile (Recommended)**
1. Add the export to your shell profile:
   ```bash
   # Add to ~/.zshrc or ~/.bashrc
   export BITRISE_TOKEN="your-actual-token-here"
   ```
2. Restart your terminal or source the profile: `source ~/.zshrc`
3. **Restart Kiro** - this is required for Kiro to pick up the new environment variable
4. When prompted, accept the popup asking if you trust the environment variable

**Option 2: Manual Configuration**
If the environment variable approach doesn't work, you can hardcode the token:
1. Open `~/.kiro/settings/mcp.json` (user level) or `.kiro/settings/mcp.json` (workspace level)
2. Find the Bitrise server entry
3. Replace `${BITRISE_TOKEN}` with your actual token value
4. Save the file and restart Kiro

**Note on Environment Variable Syntax**
The syntax for environment variables differs between Kiro CLI and IDE:
- **Kiro CLI**: Use `${env:BITRISE_TOKEN}` (with `env:` prefix)
- **Kiro IDE**: Use `${BITRISE_TOKEN}` (without prefix)

This inconsistency is a known issue. The Power is configured with `${BITRISE_TOKEN}` which works for the IDE.

### Power Not Activating
- Ensure you've entered the correct repository URL with `/tree/main/power` path
- Check that your `BITRISE_TOKEN` is valid
- Try mentioning explicit keywords like "bitrise" in your conversation

### Authentication Issues
- Verify your Personal Access Token is still valid
- Check token permissions in your Bitrise account settings
- Regenerate the token if necessary

### Connection Problems
- The power connects to `https://mcp.bitrise.io`
- Ensure you have internet connectivity
- Check if there are any firewall restrictions

## Additional Resources

- [Bitrise API Documentation](https://devcenter.bitrise.io/api/api-index/)
- [Kiro Powers Documentation](https://kiro.dev/docs/powers/)
- [MCP Protocol Documentation](https://modelcontextprotocol.io/)