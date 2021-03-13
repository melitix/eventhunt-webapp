import pytest
from selenium import webdriver

@pytest.mark.parametrize(
        "browser_choice",
        ["chrome", "firefox"])
def test_first(browser_choice):

    if browser_choice == "chrome":
        browser = webdriver.Chrome()
    else:
        browser = webdriver.Firefox()

    browser.get( 'http://localhost:8100/' )

    assert browser.title == "R// Garage"

    browser.quit()

@pytest.mark.parametrize(
        "browser_choice",
        ["chrome", "firefox"])
def test_second(browser_choice):

    if browser_choice == "chrome":
        browser = webdriver.Chrome()
    else:
        browser = webdriver.Firefox()

    browser.get( 'http://localhost:8100/' )

    # Test homepage as a guest user
    try:
        browser.find_element("xpath", "//*[text()='Log in to R// Garage']")
    except:
        assert False

    assert True

    browser.quit()
