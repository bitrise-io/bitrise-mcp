from mcp.server.fastmcp import FastMCP
from starlette.middleware.base import BaseHTTPMiddleware
from starlette.applications import Starlette
from mcp.server.fastmcp.server import MCPTool


class FastMCPWithContextVar(FastMCP):
    def __init__(self, name, request_var, *args, **kwargs):
        super().__init__(name, *args, **kwargs)
        self.request_var = request_var

    def sse_app(self) -> Starlette:
        app = super().sse_app()

        class RequestContextMiddleware(BaseHTTPMiddleware):
            async def dispatch(inner_self, request, call_next):
                context_token = self.request_var.set(request)
                try:
                    response = await call_next(request)
                finally:
                    self.request_var.reset(context_token)
                return response

        app.add_middleware(RequestContextMiddleware)
        return app

    async def list_tools(self):
        tools = await super().list_tools()

        request = self.request_var.get()
        if request:
            enabled_api_groups_header = request.headers.get(
                "x-bitrise-enabled-api-groups"
            )
            if enabled_api_groups_header:
                allowed_tools = set(
                    tool.strip() for tool in enabled_api_groups_header.split(",")
                )
                tools = [tool for tool in tools if tool in allowed_tools]
                print(f">>> Filtered tools: {[tool.name for tool in tools]}")

        return tools
