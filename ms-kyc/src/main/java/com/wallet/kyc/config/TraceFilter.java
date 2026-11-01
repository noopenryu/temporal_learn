package com.wallet.kyc.config;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.slf4j.MDC;
import org.springframework.core.Ordered;
import org.springframework.core.annotation.Order;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import java.io.IOException;
import java.security.SecureRandom;
import java.util.HexFormat;

@Component
@Order(Ordered.HIGHEST_PRECEDENCE)
public class TraceFilter extends OncePerRequestFilter {

    private final SecureRandom random = new SecureRandom();
    private final HexFormat hexFormat = HexFormat.of();

    private String generateTraceId() {
        byte[] bytes = new byte[16];
        random.nextBytes(bytes);
        return hexFormat.formatHex(bytes);
    }

    private String generateSpanId() {
        byte[] bytes = new byte[8];
        random.nextBytes(bytes);
        return hexFormat.formatHex(bytes);
    }

    @Override
    protected void doFilterInternal(HttpServletRequest request, HttpServletResponse response, FilterChain filterChain)
            throws ServletException, IOException {
        String traceId = null;
        String parentSpanId = null;

        // 1. Check W3C traceparent header: 00-<32hex>-<16hex>-<flags>
        String traceparent = request.getHeader("traceparent");
        if (traceparent != null && !traceparent.isBlank()) {
            String[] parts = traceparent.trim().split("-");
            if (parts.length >= 4 && "00".equals(parts[0]) && parts[1].length() == 32 && parts[2].length() == 16) {
                traceId = parts[1];
                parentSpanId = parts[2];
            }
        }

        // 2. Check X-Trace-Id / X-Request-Id fallback
        if (traceId == null || traceId.isBlank()) {
            traceId = request.getHeader("X-Trace-Id");
            if (traceId == null || traceId.isBlank()) {
                traceId = request.getHeader("X-Request-Id");
            }
            if (traceId != null && !traceId.isBlank()) {
                String clean = traceId.replace("-", "");
                if (clean.length() >= 32) {
                    traceId = clean.substring(0, 32);
                } else {
                    traceId = String.format("%-32s", clean).replace(' ', '0');
                }
            } else {
                traceId = generateTraceId();
            }
        }

        if (parentSpanId == null) {
            parentSpanId = request.getHeader("X-Span-Id");
        }

        String spanId = generateSpanId();

        MDC.put("trace_id", traceId);
        MDC.put("span_id", spanId);
        if (parentSpanId != null && !parentSpanId.isBlank()) {
            MDC.put("parent_span_id", parentSpanId);
        }

        response.setHeader("X-Trace-Id", traceId);
        response.setHeader("X-Span-Id", spanId);
        response.setHeader("traceparent", String.format("00-%s-%s-01", traceId, spanId));

        try {
            filterChain.doFilter(request, response);
        } finally {
            MDC.remove("trace_id");
            MDC.remove("span_id");
            MDC.remove("parent_span_id");
            MDC.remove("customer_id");
        }
    }
}
