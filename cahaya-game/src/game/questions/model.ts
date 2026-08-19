export type Question = {
  id: string;
  category: string;
  subject: string;
  grade: string;
  difficulty: string;
  question: string;
  optionA: string;
  optionB: string;
  optionC: string;
  optionD: string;
  correctAnswer: string;
  explanation: string;
  active: boolean;
};

export const QUESTION_BANK: Question[] = [];
