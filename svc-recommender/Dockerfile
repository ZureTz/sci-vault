FROM astral/uv:python3.14-alpine

WORKDIR /app

ENV PYTHONUNBUFFERED=1 \
    UV_COMPILE_BYTECODE=1 \
    UV_SYNC_TIMEOUT=180

# LibreOffice headless powers DOCX/PPTX/XLSX -> PDF conversion in the
# enrichment pipeline; font-noto-cjk ensures Chinese characters render so
# converted PDFs are still useful inputs to Gemini.
RUN apk add --no-cache libreoffice font-noto-cjk

COPY pyproject.toml uv.lock ./
RUN uv sync --frozen --no-dev

COPY . .

EXPOSE 50051

CMD ["uv", "run", "main.py"]
