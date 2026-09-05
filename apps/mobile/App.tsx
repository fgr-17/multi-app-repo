import { fetchHello } from "@hola/api-client";
import { StatusBar } from "expo-status-bar";
import { useEffect, useState } from "react";
import { Platform, StyleSheet, Text, View } from "react-native";

function resolveApiUrl(): string {
  if (process.env.EXPO_PUBLIC_API_URL) {
    return process.env.EXPO_PUBLIC_API_URL;
  }
  // Android emulator no ve localhost de la máquina host.
  if (Platform.OS === "android") {
    return "http://10.0.2.2:8080";
  }
  return "http://localhost:8080";
}

export default function App() {
  const apiUrl = resolveApiUrl();
  const [name, setName] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchHello(apiUrl)
      .then((data) => setName(data.name))
      .catch((err: Error) => setError(err.message));
  }, [apiUrl]);

  return (
    <View style={styles.container}>
      <Text style={styles.eyebrow}>Mobile · iOS / Android</Text>
      <Text style={styles.title}>{name ? `Hola ${name}` : "Hola …"}</Text>
      <Text style={styles.status}>
        {error
          ? `no pude hablar con el API: ${error}`
          : name
            ? `nombre servido por ${apiUrl}/api/hello`
            : "pidiendo el nombre al API de Go…"}
      </Text>
      <StatusBar style="dark" />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#f4efe6",
    justifyContent: "center",
    paddingHorizontal: 32,
  },
  eyebrow: {
    fontSize: 12,
    letterSpacing: 1.6,
    textTransform: "uppercase",
    color: "#6b645c",
    marginBottom: 12,
  },
  title: {
    fontSize: 56,
    lineHeight: 60,
    color: "#1c1916",
    fontWeight: "400",
  },
  status: {
    marginTop: 20,
    fontSize: 14,
    color: "#6b645c",
  },
});
