from typing import Any
from mcp.server.fastmcp import FastMCP
from starlette.middleware.base import BaseHTTPMiddleware
from starlette.applications import Starlette
import contextvars


class FastMCPWithContextVar(FastMCP):
    def __init__(
        self,
        name=None,
        tools_by_groups={},
        instructions=None,
        auth_server_provider=None,
        event_store=None,
        **settings: Any,
    ):
        super().__init__(
            name, instructions, auth_server_provider, event_store, **settings
        )
        self.tools_by_groups = tools_by_groups
        self.request_var = contextvars.ContextVar("request_var", default=None)

    def get_request(self):
        """Get the current request from the context variable."""
        return self.request_var.get()

    def streamable_http_app(self) -> Starlette:
        app = super().streamable_http_app()

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

        if request is None:
            return tools

        enabled_api_groups_header = request.headers.get("x-bitrise-enabled-api-groups")

        if enabled_api_groups_header:
            allowed_groups = set(
                tool.strip() for tool in enabled_api_groups_header.split(",")
            )
            filtered_tools = []
            for group in allowed_groups:
                group_tools = self.tools_by_groups.get(group, [])
                filtered_tools.extend(group_tools)
            tools = [tool for tool in tools if tool.name in filtered_tools]

        return tools
