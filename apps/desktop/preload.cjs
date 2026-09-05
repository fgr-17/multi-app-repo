const { contextBridge } = require("electron");

contextBridge.exposeInMainWorld("hola", {
  apiUrl: process.env.API_URL || "http://localhost:8080",
});
