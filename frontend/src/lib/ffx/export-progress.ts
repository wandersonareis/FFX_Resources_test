/**
 * Guard de progresso do export (singleton imperativo de módulo).
 *
 * O flag "um export está rodando agora" é escrito por quem dispara o export
 * (runExport, no export-import-controls) e lido pelo handler do evento
 * Progress (no progress-notifier), que só atualiza o toast de export quando
 * um export está ativo — evita pegar progresso de carga de árvore alheia.
 *
 * É lido apenas dentro de callback de evento (não em render), então não
 * precisa de reatividade: um objeto mutável de módulo basta. Mesmo espírito
 * dos singletons tabStorage/exportSelection — sem props atravessando
 * componentes.
 */
export const exportProgress = {
  active: false,

  /** Marca o início de um export (o toast passa a acompanhar Progress). */
  begin(): void {
    this.active = true;
  },

  /** Marca o fim (sucesso ou erro) — o toast deixa de ser atualizado. */
  end(): void {
    this.active = false;
  },

  /** true enquanto um export roda. */
  isActive(): boolean {
    return this.active;
  },
};
