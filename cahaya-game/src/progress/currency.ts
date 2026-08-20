// Client never chooses coin amounts. Server is authoritative.
export const CurrencyService = {
  display(n: number): string {
    return `🪙 ${Number(n || 0).toLocaleString("id-ID")}`;
  },
};
