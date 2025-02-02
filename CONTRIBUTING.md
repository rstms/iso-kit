# Contributing to iso-kit

Thank you for considering contributing to **iso-kit**! iso-kit is a Golang library for creating, inspecting, extracting, and manipulating ISO images in general. Currently, it supports ISO9660 and UDF image types. We welcome contributions of all kinds—from bug reports and feature requests to code and documentation improvements.

By participating in this project, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md).

---

## How to Contribute

There are several ways you can help:

### Reporting Bugs

If you encounter any issues or have found a bug:
- **Open an Issue:**  
  Go to [GitHub Issues](https://github.com/bgrewell/iso-kit/issues) and create a new issue.
- **Include Details:**  
  Provide clear steps to reproduce the problem, the expected vs. actual behavior, and any relevant environment details (e.g., OS, Go version).

### Suggesting Enhancements

If you have an idea for a new feature or an improvement:
- **Open an Issue:**  
  Clearly describe your suggestion and why you think it would be beneficial.
- **Label Appropriately:**  
  Use labels such as `enhancement` or `feature request` to help us categorize your suggestion.

### Submitting Pull Requests

We welcome pull requests (PRs) for bug fixes, new features, or documentation improvements. Before you begin, please review these guidelines:

#### Pull Request Workflow

1. **Fork the Repository:**

   Click the "Fork" button on [github.com/bgrewell/iso-kit](https://github.com/bgrewell/iso-kit) to create your own copy of the repository.

2. **Clone Your Fork:**

   ```bash
   git clone https://github.com/<your-username>/iso-kit.git
   cd iso-kit
   ```

3. **Create a New Branch:**

   Create a descriptive branch for your changes:

   ```bash
   git checkout -b feature/your-feature-name
   ```

4. **Make Your Changes:**

   - For code contributions, follow the existing code style and best practices for Go.
   - If you add new functionality or change existing behavior, please update or add tests.
   - Update the documentation if necessary.

5. **Run Tests:**

   Before committing your changes, make sure all tests pass:

   ```bash
   go test ./...
   ```

6. **Commit Your Changes:**

   Write clear, concise commit messages. For example:

   ```bash
   git commit -m "Add support for [feature] in iso-kit"
   ```

7. **Push Your Branch:**

   ```bash
   git push origin feature/your-feature-name
   ```

8. **Open a Pull Request:**

   Navigate to [github.com/bgrewell/iso-kit/pulls](https://github.com/bgrewell/iso-kit/pulls) and open a new PR from your branch against the `main` branch. Please include a clear description of your changes.

---

## Coding Guidelines

- **Language:**  
  This project is written in Go. Please follow idiomatic Go practices and style conventions.
  
- **Code Style:**  
  Ensure your code is formatted with `gofmt` and adheres to any existing linting rules.
  
- **Testing:**  
  Add tests for any new functionality or bug fixes. Run tests locally using:
  ```bash
  go test ./...
  ```

- **Documentation:**  
  Update or add documentation as needed. Inline code comments and README updates help others understand your contributions.

---

## Communication

If you have questions or need assistance:
- **Open an Issue:**  
  We use GitHub Issues for bug reports and feature requests.
- **Discussion:**  
  If applicable, join our community chat or mailing list (provide link/instructions here).

---

## Code of Conduct

Please review and adhere to our [Code of Conduct](CODE_OF_CONDUCT.md). We expect everyone to maintain a respectful and constructive environment.

---

## Licensing

By contributing to iso-kit, you agree that your contributions will be licensed under the same license as the project. Please see the [LICENSE](LICENSE) file for details.

---

Thank you for your interest in improving iso-kit. We appreciate your contributions and look forward to your pull requests!
