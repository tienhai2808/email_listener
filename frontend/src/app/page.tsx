import Auth from "@/components/Auth";

export default function Home() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-zinc-50 font-sans dark:bg-black">
      <main className="flex min-h-screen w-full max-w-3xl items-center justify-center py-32 px-16 bg-white dark:bg-black">
        <Auth />
      </main>
    </div>
  );
}
