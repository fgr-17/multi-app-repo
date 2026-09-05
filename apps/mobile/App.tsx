import {
  GreetingSync,
  statusLabel,
  wrapKvStore,
  type GreetingSnapshot,
} from "@hola/api-client";
import AsyncStorage from "@react-native-async-storage/async-storage";
import { StatusBar } from "expo-status-bar";
import { useEffect, useMemo, useState } from "react";
import {
  Platform,
  Pressable,
  StyleSheet,
  Text,
  TextInput,
  View,
} from "react-native";

function resolveApiUrl(): string {
  if (process.env.EXPO_PUBLIC_API_URL) {
    return process.env.EXPO_PUBLIC_API_URL;
  }
  if (Platform.OS === "android") {
    return "http://10.0.2.2:8080";
  }
  return "http://localhost:8080";
}

const empty: GreetingSnapshot = {
  greeting: null,
  online: false,
  status: "offline",
  error: null,
};

export default function App() {
  const apiUrl = resolveApiUrl();
  const store = useMemo(() => wrapKvStore(AsyncStorage, "hola-mobile:"), []);
  const sync = useMemo(() => new GreetingSync({ apiUrl, store }), [apiUrl, store]);
  const [snap, setSnap] = useState<GreetingSnapshot>(empty);
  const [draft, setDraft] = useState("");
  const name = snap.greeting?.name;

  useEffect(() => {
    const unsub = sync.subscribe(setSnap);
    void sync.start();
    return () => {
      unsub();
      sync.stop();
    };
  }, [sync]);

  return (
    <View style={styles.container}>
      <Text style={styles.eyebrow}>Mobile · iOS / Android</Text>
      <Text style={styles.title}>{name ? `Hola ${name}` : "Hola …"}</Text>
      <Text style={styles.status}>
        {snap.error ? `${statusLabel(snap.status)}: ${snap.error}` : statusLabel(snap.status)}
      </Text>
      <TextInput
        value={draft}
        onChangeText={setDraft}
        placeholder={name ?? "nombre"}
        placeholderTextColor="#8a837a"
        style={styles.input}
      />
      <Pressable
        onPress={() => {
          void sync.setName(draft || name || "");
          setDraft("");
        }}
        style={styles.button}
      >
        <Text style={styles.buttonLabel}>Guardar</Text>
      </Pressable>
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
  input: {
    marginTop: 24,
    borderWidth: 1,
    borderColor: "#d9d0c3",
    backgroundColor: "#fff",
    borderRadius: 8,
    paddingHorizontal: 12,
    paddingVertical: 10,
    fontSize: 16,
    color: "#1c1916",
  },
  button: {
    marginTop: 12,
    alignSelf: "flex-start",
    backgroundColor: "#1c1916",
    borderRadius: 8,
    paddingHorizontal: 16,
    paddingVertical: 10,
  },
  buttonLabel: {
    color: "#f4efe6",
    fontSize: 14,
  },
});
