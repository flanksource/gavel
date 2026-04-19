import Hero from "@/components/site/Hero";
import BentoGrid from "@/components/site/BentoGrid";
import FeatureList from "@/components/site/FeatureList";
import PricingSection from "@/components/site/PricingSection";
import FAQ from "@/components/site/FAQ";
import WaitlistForm from "@/components/site/WaitlistForm";

export default function Landing() {
  return (
    <>
      <Hero />
      <BentoGrid />
      <FeatureList />
      <PricingSection />
      <section className="mx-auto max-w-3xl px-6 pb-20">
        <WaitlistForm />
      </section>
      <FAQ />
    </>
  );
}
