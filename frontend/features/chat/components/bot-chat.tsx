"use client";

import { useEffect, useRef, useState } from "react";
import { browserFetch } from "@/lib/api/browser-fetch";

type Message = {
  id: string;
  role: "user" | "bot";
  content: string;
};

export function BotChat() {
  const [isOpen, setIsOpen] = useState(false);
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [sessionId, setSessionId] = useState<string>("");
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // Initialize session and welcome message
  useEffect(() => {
    if (isOpen && messages.length === 0) {
      setSessionId(crypto.randomUUID());
      setMessages([
        {
          id: crypto.randomUUID(),
          role: "bot",
          content: "Halo! Saya asisten pintar dari PAL Property. Ada yang bisa saya bantu terkait informasi atau pencarian properti?",
        },
      ]);
    }
  }, [isOpen, messages.length]);

  // Auto-scroll to bottom
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages, isLoading]);

  const handleSend = async (e?: React.FormEvent) => {
    e?.preventDefault();
    if (!input.trim() || isLoading) return;

    const userMessage: Message = {
      id: crypto.randomUUID(),
      role: "user",
      content: input.trim(),
    };

    setMessages((prev) => [...prev, userMessage]);
    setInput("");
    setIsLoading(true);

    try {
      const response = await browserFetch<any>("/api/chat/messages", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          session_id: sessionId,
          message: userMessage.content,
          filters: {},
          max_documents: 5,
        }),
      });

      if (response.data?.answer) {
        setMessages((prev) => [
          ...prev,
          {
            id: crypto.randomUUID(),
            role: "bot",
            content: response.data.answer,
          },
        ]);
      } else {
        throw new Error(response.message || "Failed to get response");
      }
    } catch (error) {
      console.error("Chat error:", error);
      setMessages((prev) => [
        ...prev,
        {
          id: crypto.randomUUID(),
          role: "bot",
          content: "Maaf, terjadi kesalahan saat menghubungi server. Silakan coba lagi.",
        },
      ]);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="fixed bottom-6 right-6 z-50 flex flex-col items-end">
      {isOpen && (
        <div className="mb-4 flex h-[500px] w-full max-w-[350px] sm:w-[380px] sm:max-w-none flex-col overflow-hidden rounded-2xl border border-border bg-card shadow-xl transition-all">
          {/* Header */}
          <div className="flex items-center justify-between border-b border-border bg-muted/30 px-5 py-4">
            <div className="flex items-center gap-3">
              <div className="flex h-8 w-8 items-center justify-center rounded-full bg-primary text-primary-foreground">
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M12 8V4H8" />
                  <rect width="16" height="12" x="4" y="8" rx="2" />
                  <path d="M2 14h2" />
                  <path d="M20 14h2" />
                  <path d="M15 13v2" />
                  <path d="M9 13v2" />
                </svg>
              </div>
              <div>
                <h3 className="text-sm font-bold text-foreground">PAL Bot</h3>
                <p className="text-xs text-muted-foreground" style={{ fontFamily: "var(--font-mono)" }}>ONLINE</p>
              </div>
            </div>
            <button
              onClick={() => setIsOpen(false)}
              className="rounded-full p-2 text-muted-foreground transition-colors hover:bg-muted"
              aria-label="Close chat"
            >
              <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M18 6 6 18" />
                <path d="m6 6 12 12" />
              </svg>
            </button>
          </div>

          {/* Messages Area */}
          <div className="flex-1 overflow-y-auto p-5 space-y-4 bg-accent/10">
            {messages.map((msg) => (
              <div
                key={msg.id}
                className={`flex w-full ${msg.role === "user" ? "justify-end" : "justify-start"}`}
              >
                <div
                  className={`max-w-[85%] rounded-2xl px-4 py-2.5 text-sm ${
                    msg.role === "user"
                      ? "bg-primary text-primary-foreground rounded-br-sm"
                      : "bg-background border border-border text-foreground rounded-bl-sm shadow-sm"
                  }`}
                >
                  {msg.content}
                </div>
              </div>
            ))}
            {isLoading && (
              <div className="flex w-full justify-start">
                <div className="max-w-[85%] rounded-2xl rounded-bl-sm border border-border bg-background px-4 py-3 shadow-sm">
                  <div className="flex space-x-1.5">
                    <div className="h-2 w-2 animate-bounce rounded-full bg-muted-foreground/40" style={{ animationDelay: "0ms" }}></div>
                    <div className="h-2 w-2 animate-bounce rounded-full bg-muted-foreground/40" style={{ animationDelay: "150ms" }}></div>
                    <div className="h-2 w-2 animate-bounce rounded-full bg-muted-foreground/40" style={{ animationDelay: "300ms" }}></div>
                  </div>
                </div>
              </div>
            )}
            <div ref={messagesEndRef} />
          </div>

          {/* Input Area */}
          <form onSubmit={handleSend} className="border-t border-border bg-background p-3">
            <div className="flex items-center gap-2 overflow-hidden rounded-full border border-border bg-muted/30 px-2 py-1 focus-within:ring-2 focus-within:ring-primary focus-within:ring-offset-1 focus-within:ring-offset-background">
              <input
                type="text"
                value={input}
                onChange={(e) => setInput(e.target.value)}
                placeholder="Ketik pesan Anda..."
                className="flex-1 bg-transparent px-3 py-2 text-sm text-foreground focus:outline-none placeholder:text-muted-foreground"
                disabled={isLoading}
              />
              <button
                type="submit"
                disabled={!input.trim() || isLoading}
                className="flex h-8 w-8 items-center justify-center rounded-full bg-primary text-primary-foreground transition-all disabled:opacity-50 hover:bg-primary/90"
              >
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                  <path d="m5 12 7-7 7 7" />
                  <path d="M12 19V5" />
                </svg>
              </button>
            </div>
          </form>
        </div>
      )}

      {/* Floating Button */}
      {!isOpen && (
        <button
          onClick={() => setIsOpen(true)}
          className="group flex h-14 w-14 items-center justify-center rounded-full bg-primary text-primary-foreground shadow-lg transition-transform hover:scale-105 active:scale-95"
          aria-label="Open chatbot"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="transition-transform group-hover:-translate-y-1">
            <path d="M7.9 20A9 9 0 1 0 4 16.1L2 22Z" />
            <path d="M8 12h.01" />
            <path d="M12 12h.01" />
            <path d="M16 12h.01" />
          </svg>
        </button>
      )}
    </div>
  );
}
