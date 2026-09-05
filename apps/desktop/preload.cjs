const { contextBridge } = require("electron");

contextBridge.exposeInMainWorld("holaEnv", {
  apiUrl: process.env.API_URL || "http://localhost:8080",
});
