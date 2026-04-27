import DOMPurify from 'dompurify'
import MarkdownIt from 'markdown-it'

const markdown = new MarkdownIt({
  html: false,
  linkify: true,
  breaks: true
})

export function renderMarkdown(markdownText = '') {
  const rawHtml = markdown.render(markdownText)
  return DOMPurify.sanitize(rawHtml, {
    USE_PROFILES: { html: true }
  })
}
