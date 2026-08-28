"""启动入口：确保 Windows 事件循环兼容性"""
import asyncio
import sys

if sys.platform == "win32":
    asyncio.set_event_loop_policy(asyncio.WindowsSelectorEventLoopPolicy())

import uvicorn
from main import app

uvicorn.run(app, host="127.0.0.1", port=8001)
