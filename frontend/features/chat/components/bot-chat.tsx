"use client";

import Link from "next/link";
import Image from "next/image";
import { useEffect, useRef, useState } from "react";
import ReactMarkdown, { type Components } from "react-markdown";
import { browserFetch } from "@/lib/api/browser-fetch";

type Recommendation = {
  listing_id: string;
  title: string;
  slug: string;
  price: number;
  currency: string;
  location_city?: string | null;
  location_district?: string | null;
  location_province?: string | null;
  primary_image_url?: string | null;
  category?: {
    id: string;
    name: string;
    slug: string;
  } | null;
};

type Message = {
  id: string;
  role: "user" | "bot";
  content: string;
  answer_format?: "text" | "markdown";
  recommendations?: Recommendation[];
};

type ChatMessageResponse = {
  answer: string;
  answer_format?: "text" | "markdown";
  recommendations?: Recommendation[];
};

const ALLOWED_MARKDOWN_ELEMENTS = ["p", "br", "h3", "h4", "strong", "em", "ul", "ol", "li", "a"] as const;
const SAFE_LISTING_HREF_PATTERN = /^\/listings\/[a-z0-9]+(?:-[a-z0-9]+)*$/;

function isSafeListingHref(href?: string): href is `/listings/${string}` {
  return typeof href === "string" && SAFE_LISTING_HREF_PATTERN.test(href);
}

const markdownComponents: Components = {
  p: ({ node: _node, ...props }) => <p className="leading-7 [&:not(:first-child)]:mt-3" {...props} />,
  h3: ({ node: _node, ...props }) => <h3 className="mt-4 text-sm font-semibold tracking-[0.02em] text-foreground" {...props} />,
  h4: ({ node: _node, ...props }) => <h4 className="mt-4 text-sm font-semibold text-foreground/90" {...props} />,
  strong: ({ node: _node, ...props }) => <strong className="font-semibold text-foreground" {...props} />,
  em: ({ node: _node, ...props }) => <em className="italic" {...props} />,
  ul: ({ node: _node, ...props }) => <ul className="mt-3 list-disc space-y-1.5 pl-5" {...props} />,
  ol: ({ node: _node, ...props }) => <ol className="mt-3 list-decimal space-y-1.5 pl-5" {...props} />,
  li: ({ node: _node, ...props }) => <li className="pl-1 leading-7" {...props} />,
  a: ({ node: _node, href, children, ...props }) => {
    if (!isSafeListingHref(href)) {
      return <span className="font-semibold text-foreground">{children}</span>;
    }

    return (
      <Link className="font-semibold text-primary underline decoration-primary/50 underline-offset-4 transition-colors hover:text-primary/80" href={href} {...props}>
        {children}
      </Link>
    );
  },
};

function formatRecommendationPrice(price: number, currency: string) {
  if ((currency || "IDR") === "IDR") {
    if (price >= 1_000_000_000) {
      return `Rp ${new Intl.NumberFormat("id-ID", { maximumFractionDigits: 1 }).format(price / 1_000_000_000)} Miliar`;
    }
    if (price >= 1_000_000) {
      return `Rp ${new Intl.NumberFormat("id-ID", { maximumFractionDigits: 1 }).format(price / 1_000_000)} Juta`;
    }
  }

  return new Intl.NumberFormat("id-ID", { style: "currency", currency: currency || "IDR", maximumFractionDigits: 0 }).format(price);
}

export function BotChat() {
  const [isOpen, setIsOpen] = useState(false);
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [sessionId, setSessionId] = useState<string>("");
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const messageCount = messages.length;

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
    if (messageCount > 0 || isLoading) {
      messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
    }
  }, [messageCount, isLoading]);

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
      const response = await browserFetch<ChatMessageResponse>("/api/chat/messages", {
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
            answer_format: response.data.answer_format === "markdown" ? "markdown" : "text",
            recommendations: response.data.recommendations ?? [],
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
                <svg aria-hidden="true" xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
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
              type="button"
            >
              <svg aria-hidden="true" xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
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
                  {msg.role === "bot" && msg.answer_format === "markdown" ? (
                    <div className="space-y-0 text-sm text-foreground">
                      <ReactMarkdown
                        allowedElements={[...ALLOWED_MARKDOWN_ELEMENTS]}
                        components={markdownComponents}
                        skipHtml
                        unwrapDisallowed
                      >
                        {msg.content}
                      </ReactMarkdown>
                    </div>
                  ) : (
                    <p className="whitespace-pre-line leading-7">{msg.content}</p>
                  )}

                  {msg.role === "bot" && msg.recommendations && msg.recommendations.length > 0 ? (
                    <div className="mt-4 space-y-3 border-t border-border/70 pt-3">
                      {msg.recommendations.map((recommendation) => (
                        <Link
                          key={recommendation.listing_id}
                          className="block rounded-2xl border border-border bg-card px-3 py-3 text-foreground transition hover:border-primary/50 hover:shadow-sm"
                          href={`/listings/${recommendation.slug}`}
                        >
                          <div className="flex items-start gap-3">
                            <div className="relative h-16 w-16 shrink-0 overflow-hidden rounded-xl bg-muted">
                              {recommendation.primary_image_url ? (
                                <Image
                                  alt={recommendation.title}
                                  fill
                                  src={recommendation.primary_image_url}
                                  className="object-cover"
                                  unoptimized
                                />
                              ) : null}
                            </div>
                            <div className="min-w-0">
                              <p className="line-clamp-2 text-sm font-semibold text-foreground">{recommendation.title}</p>
                              <p className="mt-1 text-xs text-muted-foreground line-clamp-2">
                                {[recommendation.location_district, recommendation.location_city, recommendation.location_province].filter(Boolean).join(", ")}
                              </p>
                              <p className="mt-2 text-xs font-semibold text-foreground">
                                {formatRecommendationPrice(recommendation.price, recommendation.currency)}
                              </p>
                              <span className="mt-2 inline-block text-[11px] font-semibold text-primary underline underline-offset-4">Lihat detail</span>
                            </div>
                          </div>
                        </Link>
                      ))}
                    </div>
                  ) : null}
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
                <svg aria-hidden="true" xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
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
          type="button"
        >
          <svg aria-hidden="true" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="transition-transform group-hover:-translate-y-1">
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
