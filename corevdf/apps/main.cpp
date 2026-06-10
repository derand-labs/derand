#include "common.hpp"
#include "hash/sha256.hpp"
#include "vdf/prover.hpp"
#include "vdf/verifier.hpp"

#include <cxxopts.hpp>
#include <fstream>
#include <iostream>

using namespace vdf;

namespace
{
    constexpr int EXIT_RUNTIME_ERROR = 1;
    constexpr int EXIT_VERIFY_FAILED = 2;
    const char *DEFAULT_SEED = "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f";
    const char *DEFAULT_XSEED = "202122232425262728292a2b2c2d2e2f303132333435363738393a3b3c3d3e3f";

    std::string gen_proof_id(const std::vector<uint8_t> &x_seed, const uint64_t t)
    {
        std::vector<uint8_t> seed;
        seed.insert(seed.begin(), x_seed.begin(), x_seed.end());
        common::append_u64_be(seed, t);

        return common::bytes_to_hex(hash::sha256(seed));
    }

    std::string gen_proof_path(std::filesystem::path setup_dir, std::string system_id, std::string proof_id, std::string prefix)
    {
        std::filesystem::create_directories(setup_dir / "proofs" / system_id);
        return setup_dir / "proofs" / system_id / (prefix + "-" + proof_id + ".json");
    }

    template <typename... Args>
    void printTitle(Args &&...args)
    {
        std::cout << "• ";
        ((std::cout << std::forward<Args>(args)), ...);
        std::cout << std::endl;
    }

    template <typename... Args>
    void printSubTitle(Args &&...args)
    {
        std::cout << "   → ";
        ((std::cout << std::forward<Args>(args)), ...);
        std::cout << std::endl;
    }

    void printLoadingTitle(std::string title)
    {
        std::cout << "• " << title << "..." << std::flush;
    }

    void printOK(common::Clock::time_point start)
    {
        std::cout << "\033[1;32mOK\033[0m "
                  << "(" << common::elapsed_ms(start) << ")" << std::endl;
    }

    void printFAIL(common::Clock::time_point start)
    {
        std::cout << "\033[1;31mFAIL\033[0m "
                  << "(" << common::elapsed_ms(start) << ")" << std::endl;
    }

    void printProcess(int n, int target, bool end)
    {
        static thread_local std::string prev_n_s;

        std::string target_s = std::to_string(target);

        if (prev_n_s.size() > 0)
        {
            for (size_t j = 0; j < prev_n_s.size() + target_s.size() + 1; ++j)
                std::cout << "\b \b";
        }

        if (!end)
        {
            std::string n_s = std::to_string(n);
            std::cout << n_s << "/" << target_s;

            prev_n_s = n_s;
        }
        else
        {
            prev_n_s = "";
        }
        std::cout << std::flush;
    }

    static inline uint64_t splitmix64(uint64_t x)
    {
        x += 0x9e3779b97f4a7c15ULL;
        x = (x ^ (x >> 30)) * 0xbf58476d1ce4e5b9ULL;
        x = (x ^ (x >> 27)) * 0x94d049bb133111ebULL;
        return x ^ (x >> 31);
    }
    static inline void fnv1a_append_u8(
        uint64_t &h,
        uint8_t x)
    {
        h ^= x;
        h *= 0x100000001b3ULL;
    }

    static inline void fnv1a_append_mpz(
        uint64_t &h,
        const mpz_t x)
    {
        size_t count = 0;
        void *data = mpz_export(nullptr, &count, 1, 1, 1, 0, x);
        uint8_t *p = static_cast<uint8_t *>(data);
        for (size_t i = 0; i < count; ++i)
        {
            fnv1a_append_u8(h, p[i]);
        }
        free(data);
    }

    static inline uint64_t bucket_for_form(
        const classgroup::Form &f,
        uint64_t num_buckets)
    {
        uint64_t h = 0xcbf29ce484222325ULL;

        fnv1a_append_mpz(h, f.a);
        fnv1a_append_u8(h, 0xff);

        fnv1a_append_mpz(h, f.b);
        fnv1a_append_u8(h, 0xfe);

        fnv1a_append_mpz(h, f.c);

        h = splitmix64(h);

        return h % num_buckets;
    }

    void print_root_help(const std::string &program_name)
    {
        std::cout << "Usage: " << program_name << " <subcommand> [options]\n\n";
        std::cout << "Subcommands:\n";
        std::cout << "  setup             Build and persist VDF system parameters\n";
        std::cout << "  prove             Evaluate and generate VDF proof\n";
        std::cout << "  verify            Verify VDF proof and emit transcript\n";
        std::cout << "  h2f-distribution  Check the hash to form distribution\n\n";
        std::cout << "Run '" << program_name << " <subcommand> --help' for subcommand options.\n";
    }

    int run_setup(int argc, char **argv)
    {
        cxxopts::Options options("setup", "Build and persist VDF system parameters");

        auto add_option = options.add_options();
        add_option("h,help", "Show help");
        add_option(
            "setup-dir",
            "Setup directory",
            cxxopts::value<std::string>()->default_value(".vdf"));
        add_option(
            "seed",
            "System seed (hex)",
            cxxopts::value<std::string>()->default_value(DEFAULT_SEED));
        add_option(
            "d-bits",
            "Discriminant bit size",
            cxxopts::value<uint16_t>()->default_value("6656"));
        add_option(
            "l-bits",
            "VDF Wesolowski challenge bits",
            cxxopts::value<uint16_t>()->default_value("128"));
        add_option(
            "limb-bits",
            "Limb bits",
            cxxopts::value<uint16_t>()->default_value("122"));
        add_option(
            "split-exp",
            "Split exponent for zkvdf recursive circuit",
            cxxopts::value<uint16_t>()->default_value("8"));
        add_option(
            "hash-to-form-nb-generators",
            "Number of fixed class group forms pre-generated and stored as the generators for the hash-to-form procedure",
            cxxopts::value<uint16_t>()->default_value("9699"));
        add_option(
            "hash-to-form-steps",
            "Number of sequential class group compositions performed during the hash-to-form process, where each step selects one element from the generators using randomness",
            cxxopts::value<uint16_t>()->default_value("26"));

        const auto result = options.parse(argc, argv);

        if (result.count("help"))
        {
            std::cout << options.help() << '\n';
            return 0;
        }

        const std::filesystem::path setup_dir = std::filesystem::path(result["setup-dir"].as<std::string>());
        const std::vector<uint8_t> seed = common::hex_to_bytes(result["seed"].as<std::string>());
        const uint16_t d_bits = result["d-bits"].as<uint16_t>();
        const uint16_t l_bits = result["l-bits"].as<uint16_t>();
        const uint16_t limb_bits = result["limb-bits"].as<uint16_t>();
        const uint16_t split_exp = result["split-exp"].as<uint16_t>();
        const uint16_t hash_to_form_nb_generators = result["hash-to-form-nb-generators"].as<uint16_t>();
        const uint16_t hash_to_form_steps = result["hash-to-form-steps"].as<uint16_t>();

        printTitle("System info");
        printSubTitle("d-bits: ", d_bits);
        printSubTitle("d-seed: ", common::bytes_to_hex(seed));
        printSubTitle("challenge-l-bits: ", l_bits);
        printSubTitle("hash-to-form generators: ", hash_to_form_nb_generators);
        printSubTitle("hash-to-form steps: ", hash_to_form_steps);

        const auto t_build_start = common::Clock::now();
        printLoadingTitle("Building system");
        const System system(
            seed,
            d_bits,
            l_bits,
            limb_bits,
            split_exp,
            hash_to_form_nb_generators,
            hash_to_form_steps);
        printOK(t_build_start);

        std::string path = system.save(setup_dir / "systems");

        printSubTitle("System ID: ", system.system_id);
        printSubTitle("Path: ", path);
        return 0;
    }

    int run_prove(int argc, char **argv)
    {
        cxxopts::Options options("prove", "Evaluate and generate VDF proof");
        auto add_option = options.add_options();
        add_option("h,help", "Show help");
        add_option(
            "setup-dir",
            "Setup directory",
            cxxopts::value<std::string>()->default_value(".vdf"));
        add_option("system", "System ID", cxxopts::value<std::string>());
        add_option(
            "x-seed",
            "Input seed (hex)", cxxopts::value<std::string>()->default_value(DEFAULT_XSEED));
        add_option(
            "t",
            "Iteration count",
            cxxopts::value<uint64_t>()->default_value("100000"));
        add_option(
            "eval-only",
            "Only evaluate without proving",
            cxxopts::value<bool>()->default_value("false"));

        const auto result = options.parse(argc, argv);

        if (result.count("help"))
        {
            std::cout << options.help() << '\n';
            return 0;
        }

        const std::filesystem::path setup_dir = std::filesystem::path(result["setup-dir"].as<std::string>());
        const std::string system_id = result["system"].as<std::string>();
        const uint64_t t = result["t"].as<uint64_t>();
        const std::vector<uint8_t> x_seed = common::hex_to_bytes(result["x-seed"].as<std::string>());
        const bool eval_only = result["eval-only"].as<bool>();

        if (x_seed.size() > 32)
            throw std::invalid_argument("x_seed must be at most 32 bytes");

        printTitle("Proving info");
        printSubTitle("x-seed: ", common::bytes_to_hex(x_seed));
        printSubTitle("number of iterations: ", t);

        const auto t_load_system_start = common::Clock::now();
        printLoadingTitle("Loading system");
        const System system = System::load(setup_dir / "systems", system_id);
        printOK(t_load_system_start);
        printSubTitle("System ID: ", system_id);

        const PublicStatement stmt{
            x_seed,
            t,
        };

        const auto t_eval_start = common::Clock::now();
        printLoadingTitle("Evaluating VDF");
        VdfOutput eval = evaluate_vdf(system, stmt);
        printOK(t_eval_start);

        if (!eval_only)
        {
            const auto t_prove_start = common::Clock::now();
            printLoadingTitle("Proving");
            VdfProof proof = prove_wesolowski(system, stmt, eval);
            printOK(t_prove_start);

            std::filesystem::create_directories(setup_dir / "proofs");
            const std::string proof_id = gen_proof_id(x_seed, t);
            const std::string proof_path = gen_proof_path(setup_dir, system_id, proof_id, "proof");

            {
                std::ofstream out(proof_path, std::ios::trunc);
                nlohmann::json j = proof;
                out << j.dump(4);
            }

            printSubTitle("Proof ID: ", proof_id);
            printSubTitle("Proof path: ", proof_path);
        }
        return 0;
    }

    int run_verify(int argc, char **argv)
    {
        cxxopts::Options options("verify", "Verify VDF proof and emit transcript");
        auto add_option = options.add_options();
        add_option("h,help", "Show help");
        add_option("system", "System ID", cxxopts::value<std::string>());
        add_option(
            "setup-dir",
            "Setup directory",
            cxxopts::value<std::string>()->default_value(".vdf"));
        add_option(
            "x-seed",
            "Input seed (hex)",
            cxxopts::value<std::string>()->default_value(DEFAULT_XSEED));
        add_option(
            "t",
            "Iteration count",
            cxxopts::value<uint64_t>()->default_value("100000"));
        add_option(
            "proof-path",
            "Explicit proof path",
            cxxopts::value<std::string>()->default_value(""));

        const auto result = options.parse(argc, argv);

        if (result.count("help"))
        {
            std::cout << options.help() << '\n';
            return 0;
        }

        const std::string system_id = result["system"].as<std::string>();
        const std::filesystem::path setup_dir = std::filesystem::path(result["setup-dir"].as<std::string>());
        const std::string explicit_proof_path = result["proof-path"].as<std::string>();
        const uint64_t t = result["t"].as<uint64_t>();
        const std::vector<uint8_t> x_seed = common::hex_to_bytes(result["x-seed"].as<std::string>());

        printTitle("Verifying info");
        printSubTitle("x-seed: ", common::bytes_to_hex(x_seed));
        printSubTitle("number of iterations: ", t);

        const auto t_load_system_start = common::Clock::now();
        printLoadingTitle("Loading system");
        const System system = System::load(setup_dir / "systems", system_id);
        printOK(t_load_system_start);
        printSubTitle("System ID: ", system_id);

        PublicStatement stmt{
            x_seed,
            t,
        };

        const auto t_load_proof_start = common::Clock::now();
        printLoadingTitle("Loading proof");

        const std::string proof_id = gen_proof_id(x_seed, t);
        std::string proof_path = explicit_proof_path;
        if (explicit_proof_path.empty())
        {
            proof_path = gen_proof_path(setup_dir, system_id, proof_id, "proof");
        }

        std::ifstream proof_stream(proof_path);
        if (!proof_stream.is_open())
            throw std::runtime_error("cannot open proof file: " + proof_path);

        nlohmann::json j;
        proof_stream >> j;
        VdfProof proof = j.get<VdfProof>();
        printOK(t_load_proof_start);
        printSubTitle("Proof ID: ", proof_id);
        printSubTitle("Proof path: ", proof_path);

        const auto t_verify_start = common::Clock::now();
        printLoadingTitle("Verifying proof");
        const vdf::VdfVerifyTranscript transcript = verify_wesolowski(system, stmt, proof);
        if (!transcript.ok)
        {
            printFAIL(t_verify_start);
        }
        else
        {
            printOK(t_verify_start);
        }

        std::string transcript_path = "transcript-" + explicit_proof_path;
        if (explicit_proof_path.empty())
        {
            transcript_path = gen_proof_path(setup_dir, system_id, proof_id, "transcript");
        }

        std::ofstream out(transcript_path, std::ios::trunc);
        j = transcript;
        out << j.dump(4);
        printSubTitle("Transcript path: ", transcript_path);

        return transcript.ok ? 0 : EXIT_VERIFY_FAILED;
    }

    int run_check_hash_to_form_uniform(int argc, char **argv)
    {
        cxxopts::Options options("htf-distribution", "Check hash-to-form distribution");
        auto add_option = options.add_options();
        add_option("h,help", "Show help");
        add_option("system-id", "System ID", cxxopts::value<std::string>());
        add_option(
            "system-dir",
            "System directory",
            cxxopts::value<std::string>()->default_value(".system"));
        add_option(
            "s,samples",
            "Number of samples",
            cxxopts::value<uint32_t>()->default_value("1048576"));
        add_option(
            "b,buckets",
            "Number of buckets",
            cxxopts::value<uint32_t>()->default_value("4096"));

        const auto result = options.parse(argc, argv);

        if (result.count("help"))
        {
            std::cout << options.help() << '\n';
            return 0;
        }

        const std::string system_id = result["system-id"].as<std::string>();
        const std::string system_dir = result["system-dir"].as<std::string>();
        const uint32_t num_samples = result["samples"].as<uint32_t>();
        const uint32_t num_buckets = result["buckets"].as<uint32_t>();

        const double expected = double(num_samples) / double(num_buckets);
        if (expected < 5.0)
        {
            std::cout << "Expected bucket count too small for chi-square test\n";
            std::cout << "Increase number of samples or decrease number of buckets and try again\n";
            return 1;
        }

        printTitle("Info");
        printSubTitle("Number of samples: ", num_samples);
        printSubTitle("Number of buckets: ", num_buckets);

        const auto t_load_system_start = common::Clock::now();
        printLoadingTitle("Loading system");
        const System system = System::load(system_dir, system_id);
        printOK(t_load_system_start);
        printSubTitle("System ID: ", system_id);

        std::vector<uint64_t> buckets(num_buckets, 0);
        std::vector<uint8_t> seed;
        seed.reserve(8);

        const auto t_generate_samples = common::Clock::now();
        printLoadingTitle("Generate samples");
        printProcess(0, 0, true); // reset process
        for (uint32_t i = 0; i < num_samples; ++i)
        {
            printProcess(i, num_samples, false);
            seed.clear();
            common::append_u32_be(seed, i);

            classgroup::Form f = vdf::hash_to_form(system, seed);

            const uint64_t b = bucket_for_form(f, num_buckets);
            ++buckets[b];
        }
        printProcess(num_samples, num_samples, true);
        printOK(t_generate_samples);

        double chi2 = 0.0;
        uint64_t min_count = buckets[0];
        uint64_t max_count = buckets[0];

        for (uint64_t x : buckets)
        {
            const double diff = double(x) - expected;
            chi2 += diff * diff / expected;
            if (x < min_count)
                min_count = x;
            if (x > max_count)
                max_count = x;
        }

        const double df = double(num_buckets - 1);
        const double normalized_chi2 = chi2 / df;

        // chi-square distribution:
        // mean     = df
        // variance = 2 * df
        // stddev   = sqrt(2 * df)
        const double z_score =
            (chi2 - df) / std::sqrt(2.0 * df);

        printTitle("Result");
        printSubTitle("expected each buckets: ", expected);
        printSubTitle("min count: ", min_count);
        printSubTitle("max count: ", max_count);
        printSubTitle("chi2: ", chi2);
        printSubTitle("normalized chi2 (ideal 0.8 to 1.2): ", normalized_chi2);
        printSubTitle("chi2 z-score (ideal -2 to 2): ", z_score);

        return 0;
    }
} // namespace

int main(int argc, char **argv)
{
    try
    {
        if (argc < 2)
        {
            print_root_help(argv[0]);
            return EXIT_RUNTIME_ERROR;
        }

        const std::string subcommand = argv[1];
        if (subcommand == "-h" || subcommand == "--help")
        {
            print_root_help(argv[0]);
            return 0;
        }

        if (subcommand == "setup")
        {
            return run_setup(argc - 1, argv + 1);
        }
        if (subcommand == "prove")
        {
            return run_prove(argc - 1, argv + 1);
        }
        if (subcommand == "verify")
        {
            return run_verify(argc - 1, argv + 1);
        }
        if (subcommand == "h2f-distribution")
        {
            return run_check_hash_to_form_uniform(argc - 1, argv + 1);
        }

        print_root_help(argv[0]);
        return EXIT_RUNTIME_ERROR;
    }
    catch (const std::exception &ex)
    {
        std::cerr << "corevdf error: " << ex.what() << "\n";
        return EXIT_RUNTIME_ERROR;
    }
}
