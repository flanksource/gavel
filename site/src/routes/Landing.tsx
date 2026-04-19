import Hero from "@/components/site/Hero";
import FeatureGrid from "@/components/site/FeatureGrid";
import PricingSection from "@/components/site/PricingSection";
import FAQ from "@/components/site/FAQ";
import WaitlistForm from "@/components/site/WaitlistForm";

export default function Landing() {
  return (
    <>
      <Hero />
      <FeatureGrid />
      <PricingSection />
      <section className="mx-auto max-w-3xl px-6 pb-20">
        <WaitlistForm />
      </section>
      <FAQ />
    </>
  );
}
